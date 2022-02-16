package main

import (
	// "regexp"
	"fmt"
	"testing"

	// "time"

	"context"

	umc "cs426.yale.edu/lab1/user_service/mock_client"
	usl "cs426.yale.edu/lab1/user_service/server_lib"
	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	sl "cs426.yale.edu/lab1/video_rec_service/server_lib"
	vmc "cs426.yale.edu/lab1/video_service/mock_client"
	vsl "cs426.yale.edu/lab1/video_service/server_lib"
)

func TestGetTopVids_Limit(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	response, err := vid_rec.GetTopVideos(context.Background(), &request)

	if err != nil {
		fmt.Println(err)
		t.Fatal("Error in GetTopVideos")
	} else if len(response.Videos) != 5 {
		t.Fatal("Returned length != 5")
	}
}

func TestGetTopVids_Limit_2(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  50, //larger than batch size
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	response, err := vid_rec.GetTopVideos(context.Background(), &request)

	if err != nil {
		fmt.Println(err)
		t.Fatal("Error in GetTopVideos")
	} else if len(response.Videos) != 50 {
		t.Fatal("Returned length != 50")
	}
}

func TestGetStats_1(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	ctx := context.Background()

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	for i := 0; i < 10; i++ {
		vid_rec.GetTopVideos(ctx, &request)
	}

	request2 := pb.GetStatsRequest{}
	response, err := vid_rec.GetStats(ctx, &request2)
	if err != nil {
		fmt.Println(err)
		t.Fatal("Error in GetStats")
	}
	if response.TotalRequests != 10 {
		t.Fatal("Incorrect total requests")
	}
}

func TestGetTopVids_Content(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	response, err := vid_rec.GetTopVideos(context.Background(), &request)

	if err != nil {
		fmt.Println(err)
		t.Fatal("Error in GetTopVideos")
	} else if response.Videos[0].VideoId != 1085 {
		t.Fatal("Wrong reccomendation")
	}
}

func TestBatchSize(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     20,
		DisableRetry:     false,
		DisableFallback:  false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	ctx := context.Background()
	// var counter int64
	// var stales int64
	// var wg sync.WaitGroup

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	for i := 0; i < 5; i++ {
		_, err := vid_rec.GetTopVideos(ctx, &request)
		if err == nil {
			t.Fatal("Batchsize should be causing errors")
		}
	}
}

func TestFail(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          1,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
		DisableFallback:  false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	ctx := context.Background()

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	for i := 0; i < 5; i++ {
		_, err := vid_rec.GetTopVideos(ctx, &request) //when failrate is not 0 the returned value would give a seg fault... some mysterious race condition exists here
		if err == nil {
			t.Fatal("Should be failing")
		}
	}
}

func TestStales(t *testing.T) {
	userclient := umc.MakeMockUserServiceClient(usl.UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	videoclient := vmc.MakeMockVideoServiceClient(vsl.VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           0,
		SleepNs:              0,
		FailureRate:          10,
		ResponseOmissionRate: 0,
		MaxBatchSize:         10,
	})
	vid_rec_opt := sl.VideoRecServiceOptions{
		UserServiceAddr:  "[::1]:8081",
		VideoServiceAddr: "[::1]:8082",
		MaxBatchSize:     10,
		DisableRetry:     false,
		DisableFallback:  false,
	}
	request := pb.GetTopVideosRequest{
		UserId: 204054,
		Limit:  5,
	}

	ctx := context.Background()
	var counter int64
	var stales int64

	vid_rec := sl.MakeVideoRecServiceServerWithMocks(vid_rec_opt, userclient, videoclient)
	for i := 0; i < 5; i++ {
		vid_rec.GetTopVideos(ctx, &request)
	}

	// request2 := inj.SetInjectionConfigRequest{Config: &inj.InjectionConfig{SleepNs: 0, FailureRate: 1, ResponseOmissionRate: 0}}
	// videoclient.SetInjectionConfig(ctx, &request2) // This test wont work

	for i := 0; i < 15; i++ {
		res, err := vid_rec.GetTopVideos(ctx, &request)
		if err != nil {
			counter += 1
		} else if res.StaleResponse == true {
			stales += 1
		}
	}

	if counter == 0 {
		t.Fatal("Should be failing some")
	}
	if stales == 0 {
		t.Fatal("Should be giving some stales")
	}
}
