package main

import (
	"context"
	"flag"
	"hash/fnv"
	"log"
	"time"

	upb "cs426.yale.edu/lab1/user_service/proto"
	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	vpb "cs426.yale.edu/lab1/video_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	videoRecServiceAddr = flag.String(
		"video-rec-service",
		"[::1]:8080",
		"The server address for the VideoRecService in the format of host:port",
	)
	userServiceAddr = flag.String(
		"user-service",
		"[::1]:8081",
		"The server address for the UserService in the format of host:port",
	)
	netId = flag.String(
		"net-id",
		"xs66",
		"YOUR NetId to generate a user id to retrieve top video recommendations for",
	)
)

func getUserId(netId string) uint64 {
	hasher := fnv.New64()
	hasher.Write([]byte(netId))
	// See user_service server for generated user id ranges
	return hasher.Sum64()%5000 + 200000
}

func serviceConn(address string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return grpc.Dial(address, opts...)
}

//////// Test client for **Part A6**
func fetchUser(userId uint64) {
	userServiceConn, err := serviceConn(*userServiceAddr)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer userServiceConn.Close()
	userServiceClient := upb.NewUserServiceClient(userServiceConn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := userServiceClient.GetUser(
		ctx,
		&upb.GetUserRequest{UserIds: []uint64{userId}},
	)
	if err != nil || len(out.Users) != 1 {
		log.Fatalf("Error retrieving user information for user (userId %d): %v\n", userId, err)
	}
	userInfo := out.Users[0]
	log.Printf(
		"This user has name %s, their email is %s, and their profile URL is %s\n",
		userInfo.Username,
		userInfo.Email,
		userInfo.ProfileUrl,
	)
}

func sendSingleTestRequest(userId uint64, client pb.VideoRecServiceClient) []*vpb.VideoInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := client.GetTopVideos(
		ctx,
		&pb.GetTopVideosRequest{UserId: userId, Limit: 5},
	)
	if err != nil {
		log.Fatalf(
			"Error retrieving video recommendations for user (userId %d): %v\n",
			userId,
			err,
		)
	}
	log.Println("Recommended videos:")
	for i, v := range out.Videos {
		log.Printf(
			"  [%d] Video id=%d, title=\"%s\", author=%s, url=%s",
			i,
			v.VideoId,
			v.Title,
			v.Author,
			v.Url,
		)
	}
	return out.Videos
}

func main() {
	flag.Parse()

	videoRecServiceConn, err := serviceConn(*videoRecServiceAddr)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer videoRecServiceConn.Close()
	videoRecServiceClient := pb.NewVideoRecServiceClient(videoRecServiceConn)

	userId := getUserId(*netId)
	log.Printf("Welcome %s! The UserId we picked for you is %d.\n\n", *netId, userId)
	fetchUser(userId)
	sendSingleTestRequest(userId, videoRecServiceClient)

	log.Println("\n\n\n")

	// two test cases
	userId = 204054
	log.Printf("Test case 1: UserId=%d", userId)
	videos := sendSingleTestRequest(userId, videoRecServiceClient)
	if len(videos) != 5 || videos[0].VideoId != 1085 || videos[1].Author != "Renee Blick" ||
		videos[2].VideoId != 1106 || videos[3].Url != "https://video-data.localhost/blob/1211" || videos[4].Title != "evil elsewhere" {
		log.Fatalf("Test case 1 FAILED! Wrong recommendations")
	}

	userId = 203584
	log.Printf("Test case 2: UserId=%d", userId)
	videos = sendSingleTestRequest(userId, videoRecServiceClient)
	if len(videos) != 5 || videos[0].VideoId != 1015 || videos[1].Author != "Belle Harber" ||
		videos[2].VideoId != 1268 || videos[3].Url != "https://video-data.localhost/blob/1040" || videos[4].Title != "charming towards" {
		log.Fatalf("Test case 2 FAILED! Wrong recommendations")
	}
	log.Println("OK: basic tests passed!")
}
