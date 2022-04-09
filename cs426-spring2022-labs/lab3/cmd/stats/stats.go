package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	videoRecServiceAddr = flag.String(
		"video-rec-service",
		"[::1]:8080",
		"The server address for the VideoRecService in the format of host:port",
	)
)

func serviceConn(address string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.Dial(address, opts...)
}

func printHeader() {
	fmt.Println(
		"now_us\ttotal_requests\ttotal_errors\tactive_requests\tuser_service_errors\tvideo_service_errors\taverage_latency_ms\tp99_latency_ms\tstale_responses",
	)
}

func printStats(r *pb.GetStatsResponse) {
	fmt.Printf(
		"%d\t%d\t%d\t%d\t%d\t%d\t%.2f\t%.2f\t%d\n",
		time.Now().UnixMicro(),
		r.TotalRequests,
		r.TotalErrors,
		r.ActiveRequests,
		r.UserServiceErrors,
		r.VideoServiceErrors,
		r.AverageLatencyMs,
		r.P99LatencyMs,
		r.StaleResponses,
	)
}

func gatherStats(client pb.VideoRecServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := client.GetStats(
		ctx,
		&pb.GetStatsRequest{},
	)
	if err != nil {
		log.Printf(
			"Error retrieving stats: %v\n",
			err,
		)
	} else {
		printStats(out)
	}
}

func main() {
	flag.Parse()

	videoRecServiceConn, err := serviceConn(*videoRecServiceAddr)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer videoRecServiceConn.Close()
	client := pb.NewVideoRecServiceClient(videoRecServiceConn)

	printHeader()
	gatherStats(client)
	for range time.Tick(1 * time.Second) {
		go gatherStats(client)
	}
}
