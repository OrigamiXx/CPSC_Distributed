package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	sl "cs426.yale.edu/lab1/video_rec_service/server_lib"
	"google.golang.org/grpc"
)

var (
	port            = flag.Int("port", 8080, "The server port")
	userServiceAddr = flag.String(
		"user-service",
		"[::1]:8081",
		"Server address for the UserService",
	)
	videoServiceAddr = flag.String(
		"video-service",
		"[::1]:8082",
		"Server address for the VideoService",
	)
	maxBatchSize = flag.Int(
		"batch-size",
		250,
		"Maximum size of batches sent to UserService and VideoService",
	)
)

func main() {
	flag.Parse()
	log.Printf(
		"starting the server with flags: --user-service=%s --video-service=%s --batch-size=%d\n",
		*userServiceAddr,
		*videoServiceAddr,
		*maxBatchSize,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := sl.MakeVideoRecServiceServer(sl.VideoRecServiceOptions{
		UserServiceAddr:  *userServiceAddr,
		VideoServiceAddr: *videoServiceAddr,
		MaxBatchSize:     *maxBatchSize,
	})
	pb.RegisterVideoRecServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
