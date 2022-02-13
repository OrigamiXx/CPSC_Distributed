package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "cs426.yale.edu/lab1/video_service/proto"
	vsl "cs426.yale.edu/lab1/video_service/server_lib"
	"google.golang.org/grpc"
)

var (
	port        = flag.Int("port", 8082, "The server port")
	seed        = flag.Int64("seed", 42, "Random seed for generating database data")
	ttlSeconds  = flag.Int64("ttl", 60, "TTL of TopVideos in seconds")
	sleepNs     = flag.Int64("sleep-ns", 0, "Injected latency on each request")
	failureRate = flag.Int64(
		"failure-rate",
		0,
		"Injected failure rate N (0 means no injection; o/w errors one in N requests",
	)
	responseOmissionRate = flag.Int64(
		"response-omission-rate",
		0,
		"Injected response omission rate N (0 means no injection; o/w errors one in N requests",
	)
	maxBatchSize = flag.Int(
		"batch-size",
		2000,
		"Maximum size of batches accepted",
	)
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterVideoServiceServer(s,
		vsl.MakeVideoServiceServer(
			vsl.VideoServiceOptions{
				Seed:                 *seed,
				TtlSeconds:           *ttlSeconds,
				SleepNs:              *sleepNs,
				FailureRate:          *failureRate,
				ResponseOmissionRate: *responseOmissionRate,
				MaxBatchSize:         *maxBatchSize,
			}),
	)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
