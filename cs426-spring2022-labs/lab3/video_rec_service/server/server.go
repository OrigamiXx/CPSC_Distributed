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
	client_pool_size = flag.Int("client-pool-size", 4, "client-pool-size")
	port             = flag.Int("port", 8080, "The server port")
	userServiceAddr  = flag.String(
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
		"starting the server with flags: --user-service=%s --video-service=%s --batch-size=%d --client_pool_size=%d\n",
		*userServiceAddr,
		*videoServiceAddr,
		*maxBatchSize,
		*client_pool_size,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server, err := sl.MakeVideoRecServiceServer(sl.VideoRecServiceOptions{
		ClientPoolSize:   *client_pool_size,
		UserServiceAddr:  *userServiceAddr,
		VideoServiceAddr: *videoServiceAddr,
		MaxBatchSize:     *maxBatchSize,
	})
	if err != nil {
		log.Fatalf("Failure during video_rec initialization: %v", err)
	}

	pb.RegisterVideoRecServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
