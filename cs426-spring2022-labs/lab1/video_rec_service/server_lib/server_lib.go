package server_lib

import (
	"context"
	"fmt"

	"sort"
	"sync"
	"time"

	ranker "cs426.yale.edu/lab1/ranker"
	umc "cs426.yale.edu/lab1/user_service/mock_client"
	upb "cs426.yale.edu/lab1/user_service/proto"
	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	vmc "cs426.yale.edu/lab1/video_service/mock_client"
	vpb "cs426.yale.edu/lab1/video_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type VideoRecServiceOptions struct {
	// Server address for the UserService"
	UserServiceAddr string
	// Server address for the VideoService
	VideoServiceAddr string
	// Maximum size of batches sent to UserService and VideoService
	MaxBatchSize int
	// If set, disable fallback to cache
	DisableFallback bool
	// If set, disable all retries
	DisableRetry bool
}

func DefaultVideoRecServiceOptions() VideoRecServiceOptions {
	return VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
		DisableFallback:  false,
	}
}

type VideoRecServiceServer struct {
	pb.UnimplementedVideoRecServiceServer
	options            VideoRecServiceOptions
	TotalRequests      uint64
	TotalErrors        uint64
	ActiveRequests     uint64
	UserServiceErrors  uint64
	VideoServiceErrors uint64
	StaleResponses     uint64
	AverageLatencyMs   float32
	mu                 sync.RWMutex
	mockuserclient     umc.MockUserServiceClient
	mockvideoclient    vmc.MockVideoServiceClient
	testing            bool
	fetching           bool
	ch                 chan []*vpb.VideoInfo
	trendings          []*vpb.VideoInfo
}

func MakeVideoRecServiceServer(options VideoRecServiceOptions) *VideoRecServiceServer {
	return &VideoRecServiceServer{
		options:            options,
		TotalRequests:      0,
		TotalErrors:        0,
		ActiveRequests:     0,
		UserServiceErrors:  0,
		VideoServiceErrors: 0,
		AverageLatencyMs:   0,
		StaleResponses:     0,
		testing:            false,
		fetching:           false,
		ch:                 make(chan []*vpb.VideoInfo, 1),
		// Add any data to initialize here
	}
}

func MakeVideoRecServiceServerWithMocks(options VideoRecServiceOptions, mockUserServiceClient *umc.MockUserServiceClient, mockVideoServiceClient *vmc.MockVideoServiceClient) *VideoRecServiceServer {
	// Implement your own logic here

	return &VideoRecServiceServer{
		options:            options,
		mockuserclient:     *mockUserServiceClient,
		mockvideoclient:    *mockVideoServiceClient,
		testing:            true,
		TotalRequests:      0,
		TotalErrors:        0,
		ActiveRequests:     0,
		UserServiceErrors:  0,
		VideoServiceErrors: 0,
		AverageLatencyMs:   0,
		StaleResponses:     0,
		fetching:           false,
		ch:                 make(chan []*vpb.VideoInfo, 1),
		// ...
	}
}

func contains(s []uint64, num uint64) bool {
	for _, v := range s {
		if v == num {
			return true
		}
	}
	return false
}

func (server *VideoRecServiceServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	return &pb.GetStatsResponse{
		TotalRequests:      server.TotalRequests,
		TotalErrors:        server.TotalErrors,
		ActiveRequests:     server.ActiveRequests,
		UserServiceErrors:  server.UserServiceErrors,
		VideoServiceErrors: server.VideoServiceErrors,
		AverageLatencyMs:   server.AverageLatencyMs * 1e-6,
		StaleResponses:     server.StaleResponses}, status.Error(codes.OK, "OK")
}

// type GetStatsResponse struct {
// 	TotalRequests      uint64  `protobuf:"varint,1,opt,name=total_requests,json=totalRequests,proto3" json:"total_requests,omitempty"`
// 	TotalErrors        uint64  `protobuf:"varint,2,opt,name=total_errors,json=totalErrors,proto3" json:"total_errors,omitempty"`
// 	ActiveRequests     uint64  `protobuf:"varint,3,opt,name=active_requests,json=activeRequests,proto3" json:"active_requests,omitempty"`
// 	UserServiceErrors  uint64  `protobuf:"varint,4,opt,name=user_service_errors,json=userServiceErrors,proto3" json:"user_service_errors,omitempty"`
// 	VideoServiceErrors uint64  `protobuf:"varint,5,opt,name=video_service_errors,json=videoServiceErrors,proto3" json:"video_service_errors,omitempty"`
// 	AverageLatencyMs   float32 `protobuf:"fixed32,6,opt,name=average_latency_ms,json=averageLatencyMs,proto3" json:"average_latency_ms,omitempty"`
// 	P99LatencyMs       float32 `protobuf:"fixed32,7,opt,name=p99_latency_ms,json=p99LatencyMs,proto3" json:"p99_latency_ms,omitempty"`    // ExtraCredit3
// 	StaleResponses     uint64  `protobuf:"varint,8,opt,name=stale_responses,json=staleResponses,proto3" json:"stale_responses,omitempty"` // For part C
// }

type videoScore struct {
	sCore uint64
	video *vpb.VideoInfo
}

func (server *VideoRecServiceServer) GetTopVideos(ctx context.Context, req *pb.GetTopVideosRequest) (*pb.GetTopVideosResponse, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	ops := server.options
	batchsize := ops.MaxBatchSize

	server.mu.Lock()
	server.ActiveRequests += 1
	server.mu.Unlock()

	start := time.Now()
	defer func() {
		end := time.Since(start)
		// fmt.Println(end)
		server.mu.Lock()
		server.ActiveRequests -= 1
		server.AverageLatencyMs = (server.AverageLatencyMs*float32(server.TotalRequests) + float32(end)) / float32(server.TotalRequests+1)
		// fmt.Println(server.AverageLatencyMs)
		server.TotalRequests += 1
		server.mu.Unlock()
	}()

	var user_conn *grpc.ClientConn
	var userClient upb.UserServiceClient
	var err error
	if server.testing {
		// fmt.Println("Using Mock UserClient")
		userClient = &server.mockuserclient
	} else {
		user_conn, err = grpc.Dial(ops.UserServiceAddr, opts...)

		if err != nil {
			if !server.options.DisableRetry {
				user_conn, err = grpc.Dial(ops.UserServiceAddr, opts...)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.UserServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			return nil, status.Error(codes.Aborted, "User Service Dial Failed")
		}
		defer user_conn.Close()
		userClient = upb.NewUserServiceClient(user_conn)
	}

	userid := req.UserId
	request1 := upb.GetUserRequest{UserIds: []uint64{userid}}
	user, err := userClient.GetUser(ctx, &request1)

	if err != nil {
		if !server.options.DisableRetry {
			user, err = userClient.GetUser(ctx, &request1)
		}
	} //retry

	if err != nil {
		server.mu.Lock()
		server.UserServiceErrors += 1
		server.TotalErrors += 1
		server.mu.Unlock()
		if e, ok := status.FromError(err); ok {
			err = status.Errorf(e.Code(), "Error at 1 user service call: %q.", e.Code())
		} else {
			err = status.Errorf(e.Code(), "Error at 1 user service call: %q.", e.Code())
		}
		return nil, err
	}

	subscribed := user.Users[0].SubscribedTo
	// fmt.Println(len(subscribed))
	coef := user.Users[0].UserCoefficients
	var liked []*upb.UserInfo

	for i := 0; i <= len(subscribed)-batchsize; i = i + batchsize {
		request2 := upb.GetUserRequest{UserIds: subscribed[i : i+batchsize]}
		users, err := userClient.GetUser(ctx, &request2)

		if err != nil {
			if !server.options.DisableRetry {
				users, err = userClient.GetUser(ctx, &request2)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.UserServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			if e, ok := status.FromError(err); ok {
				err = status.Errorf(e.Code(), "Error at 2 user service call: %q.", e.Code())
			} else {
				err = status.Errorf(e.Code(), "Error at 2 user service call: %q.", e.Code())
			}
			return nil, err
		}
		liked = append(liked, users.Users...)
	}
	if len(subscribed)%batchsize != 0 {
		request2 := upb.GetUserRequest{UserIds: subscribed[len(subscribed)-len(subscribed)%batchsize:]}
		users, err := userClient.GetUser(ctx, &request2)

		if err != nil {
			if !server.options.DisableRetry {
				users, err = userClient.GetUser(ctx, &request2)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.UserServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			if e, ok := status.FromError(err); ok {
				err = status.Errorf(e.Code(), "Error at 2 user service call: %q.", e.Code())
			} else {
				err = status.Errorf(e.Code(), "Error at 2 user service call: %q.", e.Code())
			}
			return nil, err
		}
		liked = append(liked, users.Users...)
	}

	// fmt.Println(liked)
	var vids []uint64
	for _, v := range liked {
		for _, w := range v.LikedVideos {
			if !contains(vids, w) {
				vids = append(vids, w)
			}
		}
	}
	// fmt.Println(vids)

	// var trendings []*vpb.VideoInfo
	if len(server.ch) > 0 && !server.options.DisableFallback {
		fmt.Println("Receiving trending videos")
		server.trendings = <-server.ch
	}
	if len(server.trendings) > 0 {
		fmt.Println(server.trendings[0].Title)
		fmt.Println(server.trendings[0].VideoId)
	}

	var video_conn *grpc.ClientConn
	var videoClient vpb.VideoServiceClient
	// var err error
	if server.testing {
		// fmt.Println("Using Mock VideoClient")
		videoClient = &server.mockvideoclient
	} else {
		video_conn, err = grpc.Dial(ops.VideoServiceAddr, opts...)

		if err != nil {
			if !server.options.DisableRetry {
				video_conn, err = grpc.Dial(ops.VideoServiceAddr, opts...)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.VideoServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			if !server.options.DisableFallback {
				if len(server.ch) > 0 {
					fmt.Println("Receiving trending videos")
					server.trendings = <-server.ch
				}
				if server.trendings != nil {
					out := pb.GetTopVideosResponse{Videos: server.trendings, StaleResponse: true}
					server.mu.Lock()
					server.StaleResponses += 1
					server.mu.Unlock()
					fmt.Println("Giving stale response")
					return &out, status.Error(codes.Aborted, "Video Service Dial Failed, trending videos returned")
				}
			}
			return nil, status.Error(codes.Aborted, "Video Service Dial Failed")
		}
		defer video_conn.Close()
		videoClient = vpb.NewVideoServiceClient(video_conn)
	}

	if !server.fetching && !server.options.DisableFallback {
		server.fetching = true

		go func() {
			fmt.Println("Fetching routine running...")
			var err error
			var trends *vpb.GetTrendingVideosResponse
			var obsolete uint64
			var videos []*vpb.VideoInfo
			var done bool
			obsolete = 0

			for {
				fmt.Println(obsolete)
				fmt.Println(uint64(time.Now().Unix()))
				if uint64(time.Now().Unix()) > obsolete {
					trends, err = videoClient.GetTrendingVideos(ctx, &vpb.GetTrendingVideosRequest{})
					if err != nil {
						time.Sleep(10 * time.Second)
					} else {
						fmt.Println("Fetching new trendings...")
						fmt.Println(len(trends.Videos))
						videos = nil
						done = true
						for i := 0; i <= len(trends.Videos)-batchsize; i = i + batchsize {
							vid_batch, err := videoClient.GetVideo(ctx, &vpb.GetVideoRequest{VideoIds: trends.Videos[i : i+batchsize]})
							if err != nil {
								done = false
								break
							} else {
								videos = append(videos, vid_batch.Videos...)
							}
						}
						if len(vids)%batchsize != 0 && done {
							vid_batch, err := videoClient.GetVideo(ctx, &vpb.GetVideoRequest{VideoIds: trends.Videos[len(trends.Videos)-len(trends.Videos)%batchsize:]})
							if err != nil {
								done = false
							} else {
								videos = append(videos, vid_batch.Videos...)
							}
						}
						if !done {
							time.Sleep(10 * time.Second)
						} else {
							obsolete = trends.ExpirationTimeS
							fmt.Println(len(server.ch))
							if len(server.ch) > 0 {
								fmt.Println("Clearing...")
								<-server.ch
							}
							fmt.Println("Sending trending videos...")
							server.ch <- videos
						}
					}
				}
				time.Sleep(2 * time.Second)
			}
		}()
	}

	var videos []*vpb.VideoInfo

	for i := 0; i <= len(vids)-batchsize; i = i + batchsize {
		request3 := vpb.GetVideoRequest{VideoIds: vids[i : i+batchsize]}
		vid_batch, err := videoClient.GetVideo(ctx, &request3)

		if err != nil {
			if !server.options.DisableRetry {
				vid_batch, err = videoClient.GetVideo(ctx, &request3)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.VideoServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			if e, ok := status.FromError(err); ok {
				err = status.Errorf(e.Code(), "Error at 3 video service call: %q.", e.Code())
			} else {
				err = status.Errorf(e.Code(), "Error at 3 video service call: %q.", e.Code())
			}
			if !server.options.DisableFallback {
				if len(server.ch) > 0 {
					fmt.Println("Receiving trending videos")
					server.trendings = <-server.ch
				}
				if server.trendings != nil {
					out := pb.GetTopVideosResponse{Videos: server.trendings, StaleResponse: true}
					server.mu.Lock()
					server.StaleResponses += 1
					server.mu.Unlock()
					fmt.Println("Giving stale response")
					return &out, err
				}
			}
			return nil, err
		}
		videos = append(videos, vid_batch.Videos...)
	}
	if len(vids)%batchsize != 0 {
		request3 := vpb.GetVideoRequest{VideoIds: vids[len(vids)-len(vids)%batchsize:]}
		vid_batch, err := videoClient.GetVideo(ctx, &request3)

		if err != nil {
			if !server.options.DisableRetry {
				vid_batch, err = videoClient.GetVideo(ctx, &request3)
			}
		} //retry

		if err != nil {
			server.mu.Lock()
			server.VideoServiceErrors += 1
			server.TotalErrors += 1
			server.mu.Unlock()
			if e, ok := status.FromError(err); ok {
				err = status.Errorf(e.Code(), "Error at 3 video service call: %q.", e.Code())
			} else {
				err = status.Errorf(e.Code(), "Error at 3 video service call: %q.", e.Code())
			}
			if !server.options.DisableFallback {
				if len(server.ch) > 0 {
					fmt.Println("Receiving trending videos")
					server.trendings = <-server.ch
				}
				if server.trendings != nil {
					out := pb.GetTopVideosResponse{Videos: server.trendings, StaleResponse: true}
					server.mu.Lock()
					server.StaleResponses += 1
					server.mu.Unlock()
					fmt.Println("Giving stale response")
					return &out, err
				}
			}
			return nil, err
		}
		videos = append(videos, vid_batch.Videos...)
	}

	rank := ranker.BcryptRanker{}

	var rank_vids []videoScore
	for _, v := range videos {
		score := rank.Rank(coef, v.VideoCoefficients)
		rank_vids = append(rank_vids, videoScore{sCore: score, video: v})
	}

	sort.Slice(rank_vids, func(i, j int) bool {
		return rank_vids[i].sCore > rank_vids[j].sCore // desending sort
	})
	// fmt.Println(rank_vids)
	rank_vids = rank_vids[0:req.Limit]
	var rec []*vpb.VideoInfo
	for _, v := range rank_vids {
		rec = append(rec, v.video)
	}

	out := pb.GetTopVideosResponse{Videos: rec, StaleResponse: false}

	return &out, status.Error(codes.OK, "OK")
}
