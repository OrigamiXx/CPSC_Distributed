package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	videoRecServiceAddr = flag.String(
		"video-rec-service",
		"[::1]:8080",
		"The server address for the VideoRecService in the format of host:port",
	)
	targetQps = flag.Int(
		"target-qps",
		20,
		"Target throughput (num GetTopVideosRequests per second) for the generated traffic",
	)
)

func serviceConn(address string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.Dial(address, opts...)
}

var (
	totalSent      uint64 = 0
	totalResponses uint64 = 0
	totalSuccesses uint64 = 0
	totalLatencyMs uint64 = 0
	totalStale     uint64 = 0
)

const USER_ID_OFFSET = 200000
const MIN_USER_COUNT = 5000

func getNextUserId(userId uint64) uint64 {
	return (userId+1)%MIN_USER_COUNT + USER_ID_OFFSET
}

func sendRequest(client pb.VideoRecServiceClient, userId uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	atomic.AddUint64(&totalSent, 1)
	start := time.Now()
	wasSuccess := false
	defer func() {
		atomic.AddUint64(&totalResponses, 1)
		elapsed := uint64(time.Now().Sub(start).Milliseconds())
		atomic.AddUint64(&totalLatencyMs, elapsed)
		if wasSuccess {
			atomic.AddUint64(&totalSuccesses, 1)
		}
	}()

	responseLimit := rand.Int31n(10) + 1
	response, err := client.GetTopVideos(ctx, &pb.GetTopVideosRequest{
		UserId: userId,
		Limit:  responseLimit,
	})
	if err == nil {
		wasSuccess = true
		if response.StaleResponse {
			atomic.AddUint64(&totalStale, 1)
		}
	}
}

func reportStats() {
	for range time.Tick(1 * time.Second) {
		numReqs := atomic.LoadUint64(&totalResponses)
		numErrors := numReqs - atomic.LoadUint64(&totalSuccesses)
		latencyMs := atomic.LoadUint64(&totalLatencyMs)
		avgLatencyMs := float64(latencyMs) / float64(numReqs)
		staleResponses := atomic.LoadUint64(&totalStale)
		fmt.Printf(
			"total_sent:%d\ttotal_responses:%d\ttotal_errors:%d\tfailure_rate:%.2f%%\tstale_responses:%d\tavg_latency_ms:%.2f\n",
			atomic.LoadUint64(&totalSent),
			numReqs,
			numErrors,
			float64(numErrors)/float64(numReqs)*100,
			staleResponses,
			avgLatencyMs,
		)
	}
}

func main() {
	flag.Parse()
	go reportStats()

	maxBurst := 100
	if maxBurst > *targetQps {
		maxBurst = *targetQps
	}
	limiter := rate.NewLimiter(rate.Limit(*targetQps), maxBurst)

	numConns := 1000
	if *targetQps < numConns {
		numConns = *targetQps
	}

	for i := 0; i < numConns; i++ {
		go func() {
			videoRecServiceConn, err := serviceConn(*videoRecServiceAddr)
			if err != nil {
				log.Fatalf("fail to dial: %v", err)
			}
			defer videoRecServiceConn.Close()
			client := pb.NewVideoRecServiceClient(videoRecServiceConn)
			userId := uint64(USER_ID_OFFSET)

			for {
				if limiter.Wait(context.Background()) == nil {
					go sendRequest(client, userId)
					userId = getNextUserId(userId)
				}
			}
		}()
	}

	// block until killed
	select {}
}
