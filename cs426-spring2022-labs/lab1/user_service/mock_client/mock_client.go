package mock_client

import (
	"context"

	fi "cs426.yale.edu/lab1/failure_injection/proto"
	pb "cs426.yale.edu/lab1/user_service/proto"
	usl "cs426.yale.edu/lab1/user_service/server_lib"
	grpc "google.golang.org/grpc"
)

type MockUserServiceClient struct {
	inMemServer *usl.UserServiceServer
}

func MakeMockUserServiceClient(options usl.UserServiceOptions) *MockUserServiceClient {
	return &MockUserServiceClient{
		inMemServer: usl.MakeUserServiceServer(options),
	}
}

func (client *MockUserServiceClient) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
	_ ...grpc.CallOption,
) (*pb.GetUserResponse, error) {
	return client.inMemServer.GetUser(ctx, req)
}

func (client *MockUserServiceClient) SetInjectionConfig(
	ctx context.Context,
	req *fi.SetInjectionConfigRequest,
	_ ...grpc.CallOption,
) (*fi.SetInjectionConfigResponse, error) {
	return client.inMemServer.SetInjectionConfig(ctx, req)
}
