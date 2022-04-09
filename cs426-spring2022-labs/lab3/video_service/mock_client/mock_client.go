package mock_client

import (
	"context"

	fi "cs426.yale.edu/lab1/failure_injection/proto"
	pb "cs426.yale.edu/lab1/video_service/proto"
	vsl "cs426.yale.edu/lab1/video_service/server_lib"
	grpc "google.golang.org/grpc"
)

type MockVideoServiceClient struct {
	inMemServer *vsl.VideoServiceServer
}

func MakeMockVideoServiceClient(options vsl.VideoServiceOptions) *MockVideoServiceClient {
	return &MockVideoServiceClient{
		inMemServer: vsl.MakeVideoServiceServer(options),
	}
}

func (client *MockVideoServiceClient) GetVideo(
	ctx context.Context,
	req *pb.GetVideoRequest,
	_ ...grpc.CallOption,
) (*pb.GetVideoResponse, error) {
	return client.inMemServer.GetVideo(ctx, req)
}

func (client *MockVideoServiceClient) GetTrendingVideos(
	ctx context.Context,
	req *pb.GetTrendingVideosRequest,
	_ ...grpc.CallOption,
) (*pb.GetTrendingVideosResponse, error) {
	return client.inMemServer.GetTrendingVideos(ctx, req)
}

func (client *MockVideoServiceClient) SetInjectionConfig(
	ctx context.Context,
	req *fi.SetInjectionConfigRequest,
	opts ...grpc.CallOption,
) (*fi.SetInjectionConfigResponse, error) {
	return client.inMemServer.SetInjectionConfig(ctx, req)
}
