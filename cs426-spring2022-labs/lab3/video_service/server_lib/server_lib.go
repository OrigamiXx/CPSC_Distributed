package server_lib

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cs426.yale.edu/lab1/failure_injection"
	fi "cs426.yale.edu/lab1/failure_injection/proto"
	pb "cs426.yale.edu/lab1/video_service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gofakeit "github.com/brianvoe/gofakeit/v6"
)

type VideoServiceOptions struct {
	// Random seed for generating database data
	Seed int64
	// TTL of TopVideos in seconds
	TtlSeconds int64
	// failure injection config params
	SleepNs              int64
	FailureRate          int64
	ResponseOmissionRate int64
	// Maximum size of batches accepted
	MaxBatchSize int
}

func DefaultVideoServiceOptions() *VideoServiceOptions {
	return &VideoServiceOptions{
		Seed:                 42,
		TtlSeconds:           60,
		SleepNs:              0,
		FailureRate:          0,
		ResponseOmissionRate: 0,
		MaxBatchSize:         250,
	}
}

const VIDEO_ID_OFFSET = 1000

func makeRandomVideo(videoId uint64, maxVideos int) *pb.VideoInfo {
	video := new(pb.VideoInfo)

	video.VideoId = videoId
	titleCase := rand.Intn(4)
	if titleCase == 0 {
		video.Title = gofakeit.AdjectiveDescriptive() + " " + gofakeit.Vegetable()
	} else if titleCase == 1 {
		video.Title = "The " + gofakeit.AdjectiveDescriptive() + " " + gofakeit.Animal() + "'s " + gofakeit.NounUncountable()
	} else if titleCase == 2 {
		video.Title = gofakeit.AppName() + ": " + gofakeit.HackerVerb()
	} else {
		video.Title = gofakeit.AdjectiveDescriptive() + " " + gofakeit.AdverbPlace()
	}
	video.Url = fmt.Sprintf("https://video-data.localhost/blob/%d", videoId)
	video.Author = gofakeit.Name()

	coeffCount := rand.Intn(10) + 10
	video.VideoCoefficients = new(pb.VideoCoefficients)
	video.VideoCoefficients.Coeffs = make(map[int32]uint64)
	for i := 0; i < coeffCount; i++ {
		video.VideoCoefficients.Coeffs[int32(rand.Intn(20))] = uint64(rand.Intn(500))
	}
	return video
}

func deduplicateIds(ids []uint64) []uint64 {
	deduped := make([]uint64, 0)
	set := make(map[uint64]struct{}, 0)
	for _, id := range ids {
		if _, in_set := set[id]; !in_set {
			set[id] = struct{}{}
			deduped = append(deduped, id)
		}
	}
	return deduped
}

func makeRandomVideos() (map[uint64]*pb.VideoInfo, []uint64) {
	videoCount := rand.Intn(500) + 100

	videos := make(map[uint64]*pb.VideoInfo)
	for videoId := VIDEO_ID_OFFSET; videoId < VIDEO_ID_OFFSET+videoCount; videoId++ {
		videos[uint64(videoId)] = makeRandomVideo(uint64(videoId), videoCount)
	}

	topVideos := make([]uint64, 0)
	topVideoCount := rand.Intn(10) + 10
	for i := 0; i < topVideoCount; i++ {
		topVideos = append(topVideos, uint64(rand.Intn(videoCount)+VIDEO_ID_OFFSET))
	}
	return videos, deduplicateIds(topVideos)
}

type VideoServiceServer struct {
	pb.UnimplementedVideoServiceServer
	options   VideoServiceOptions
	videos    map[uint64]*pb.VideoInfo
	topVideos []uint64
}

func MakeVideoServiceServer(options VideoServiceOptions) *VideoServiceServer {
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
	videos, topVideos := makeRandomVideos()
	return &VideoServiceServer{
		options:   options,
		videos:    videos,
		topVideos: topVideos,
	}
}

func (db *VideoServiceServer) GetVideo(
	ctx context.Context,
	req *pb.GetVideoRequest,
) (*pb.GetVideoResponse, error) {
	shouldError := failure_injection.MaybeInject()
	if shouldError {
		return nil, status.Error(
			codes.Internal,
			"VideoService: (injected) internal error!",
		)
	}
	videoIds := req.GetVideoIds()
	if videoIds == nil || len(videoIds) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"VideoService: video_ids in GetVideoRequest should not be empty",
		)
	}
	if len(videoIds) > db.options.MaxBatchSize {
		return nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf(
				"VideoService: video_ids exceeded the max batch size %d",
				db.options.MaxBatchSize,
			),
		)
	}
	if len(deduplicateIds(videoIds)) != len(videoIds) {
		return nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("VideoService: duplicate IDs found in video_ids"),
		)
	}
	videos := make([]*pb.VideoInfo, 0, len(videoIds))
	for _, videoId := range videoIds {
		info, ok := db.videos[videoId]
		if ok {
			videos = append(videos, info)
		} else {
			return nil, status.Error(codes.NotFound, fmt.Sprintf(
				"VideoService: video %d cannot be found, it may have been deleted or never existed in the first place.",
				videoId,
			))
		}
	}
	return &pb.GetVideoResponse{Videos: videos}, nil
}

func (db *VideoServiceServer) GetTrendingVideos(
	ctx context.Context,
	req *pb.GetTrendingVideosRequest,
) (*pb.GetTrendingVideosResponse, error) {
	shouldError := failure_injection.MaybeInject()
	if shouldError {
		return nil, status.Error(
			codes.Internal,
			"VideoService: (injected) internal error!",
		)
	}
	response := new(pb.GetTrendingVideosResponse)
	response.Videos = db.topVideos
	response.ExpirationTimeS = uint64(time.Now().Unix() + db.options.TtlSeconds)
	return response, nil
}

func (db *VideoServiceServer) SetInjectionConfig(
	ctx context.Context,
	req *fi.SetInjectionConfigRequest,
) (*fi.SetInjectionConfigResponse, error) {
	failure_injection.SetInjectionConfigPb(req.Config)
	return &fi.SetInjectionConfigResponse{}, nil
}
