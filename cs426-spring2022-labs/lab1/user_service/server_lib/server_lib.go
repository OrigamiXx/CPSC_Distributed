package server_lib

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"cs426.yale.edu/lab1/failure_injection"
	fi "cs426.yale.edu/lab1/failure_injection/proto"
	pb "cs426.yale.edu/lab1/user_service/proto"
	gofakeit "github.com/brianvoe/gofakeit/v6"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceOptions struct {
	// Random seed for generating database data
	Seed int64
	// failure injection config params
	SleepNs              int64
	FailureRate          int64
	ResponseOmissionRate int64
	// maximum size of batches accepted
	MaxBatchSize int
}

func DefaultUserServiceOptions() *UserServiceOptions {
	return &UserServiceOptions{
		Seed:                 42,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         250,
	}
}

const USER_ID_OFFSET = 200000
const VIDEO_ID_OFFSET = 1000

func makeRandomUser(userId uint64, maxUsers int, maxVideos int) *pb.UserInfo {
	user := new(pb.UserInfo)

	user.UserId = userId
	user.Email = gofakeit.Email()
	user.ProfileUrl = fmt.Sprintf("https://user-service.localhost/profile/%d", userId)
	user.Username = gofakeit.Username()

	coeffCount := rand.Intn(10) + 10
	user.UserCoefficients = new(pb.UserCoefficients)
	user.UserCoefficients.Coeffs = make(map[int32]uint64)
	for i := 0; i < coeffCount; i++ {
		user.UserCoefficients.Coeffs[int32(rand.Intn(20))] = uint64(rand.Intn(500))
	}

	subscribeCount := rand.Intn(50) + 10
	for i := 0; i < subscribeCount; i++ {
		subscribed := uint64(rand.Intn(maxUsers) + USER_ID_OFFSET)
		if subscribed != userId {
			user.SubscribedTo = append(user.SubscribedTo, subscribed)
		}
	}

	likeCount := rand.Intn(20) + 5
	for i := 0; i < likeCount; i++ {
		liked := uint64(rand.Intn(maxVideos) + VIDEO_ID_OFFSET)
		user.LikedVideos = append(user.LikedVideos, liked)
	}
	return user
}

func makeRandomUsers() map[uint64]*pb.UserInfo {
	videoCount := rand.Intn(500) + 100
	userCount := rand.Intn(5000) + 5000

	users := make(map[uint64]*pb.UserInfo)

	for userId := USER_ID_OFFSET; userId < USER_ID_OFFSET+userCount; userId++ {
		users[uint64(userId)] = makeRandomUser(uint64(userId), userCount, videoCount)
	}
	return users
}

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	options UserServiceOptions
	users   map[uint64]*pb.UserInfo
}

func MakeUserServiceServer(options UserServiceOptions) *UserServiceServer {
	rand.Seed(options.Seed)
	gofakeit.Seed(options.Seed)
	failure_injection.SetInjectionConfig(
		options.SleepNs,
		options.FailureRate,
		options.ResponseOmissionRate,
	)
	fiConfig := failure_injection.GetInjectionConfig()
	log.Printf(
		"failure injection config: [sleepNs: %d, failureRate: %d, responseOmissionRate: %d",
		fiConfig.SleepNs,
		fiConfig.FailureRate,
		fiConfig.ResponseOmissionRate,
	)
	return &UserServiceServer{
		users:   makeRandomUsers(),
		options: options,
	}
}

func (db *UserServiceServer) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	shouldError := failure_injection.MaybeInject()
	if shouldError {
		return nil, status.Error(
			codes.Internal,
			"UserService: (injected) internal error!",
		)
	}

	userIds := req.GetUserIds()
	if len(userIds) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"UserService: user_ids in GetUserRequest should not be empty",
		)
	}
	if len(userIds) > db.options.MaxBatchSize {
		return nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf(
				"UserService: user_ids exceeded the max batch size %d",
				db.options.MaxBatchSize,
			),
		)
	}
	users := make([]*pb.UserInfo, 0, len(userIds))
	for _, userId := range req.GetUserIds() {
		info, ok := db.users[userId]
		if ok {
			users = append(users, info)
		} else {
			return nil, status.Error(codes.NotFound, fmt.Sprintf(
				"UserService: user %d cannot be found, it may have been deleted or never existed in the first place.",
				userId,
			))
		}
	}
	return &pb.GetUserResponse{Users: users}, nil
}

func (db *UserServiceServer) SetInjectionConfig(
	ctx context.Context,
	req *fi.SetInjectionConfigRequest,
) (*fi.SetInjectionConfigResponse, error) {
	failure_injection.SetInjectionConfigPb(req.Config)
	return &fi.SetInjectionConfigResponse{}, nil
}
