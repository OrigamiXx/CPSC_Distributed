package ranker

import (
	upb "cs426.yale.edu/lab1/user_service/proto"
	vpb "cs426.yale.edu/lab1/video_service/proto"
	"math/rand"
	"runtime"
	"testing"
)

func expectEq(t *testing.T, lhs, rhs uint64) {
	if lhs != rhs {
		_, file, line, _ := runtime.Caller( /* skip */ 1)
		t.Errorf("scores %v and %v are expected to be equal. (%s:%d)", lhs, rhs, file, line)
	}
}

func expectNeq(t *testing.T, lhs, rhs uint64) {
	if lhs == rhs {
		_, file, line, _ := runtime.Caller( /* skip */ 1)
		t.Errorf("scores %v and %v are expected to be unequal. (%s:%d)", lhs, rhs, file, line)
	}
}

func makeRandomCoeffs() map[int32]uint64 {
	result := make(map[int32]uint64)
	coeffCount := rand.Intn(10) + 10
	for i := 0; i < coeffCount; i++ {
		result[int32(rand.Intn(20))] = uint64(rand.Intn(500))
	}
	return result
}

func testSimple(t *testing.T, r Ranker) {
	rand.Seed(42)
	coeffs := makeRandomCoeffs()
	userCoeffs := &upb.UserCoefficients{Coeffs: coeffs}
	videoCoeffs := &vpb.VideoCoefficients{Coeffs: makeRandomCoeffs()}

	score1 := r.Rank(userCoeffs, videoCoeffs)
	// deterministic
	score2 := r.Rank(userCoeffs, videoCoeffs)
	expectEq(t, score1, score2)

	// perturb the vectors a bit, expect different scores
	userCoeffs.GetCoeffs()[0] = (userCoeffs.GetCoeffs()[0] + 1) * 2
	videoCoeffs.GetCoeffs()[0] = (videoCoeffs.GetCoeffs()[0] + 1) * 2
	score3 := r.Rank(userCoeffs, videoCoeffs)
	expectNeq(t, score1, score3)

	userCoeffs.GetCoeffs()[0] = (userCoeffs.GetCoeffs()[0] + 1) * 2
	score4 := r.Rank(userCoeffs, videoCoeffs)
	expectNeq(t, score1, score4)
	expectNeq(t, score3, score4)

	// ensure the ranker only dependends on the value
	coeffsCopy := make(map[int32]uint64)
	for featureId, coeff := range coeffs {
		coeffsCopy[featureId] = coeff
	}
	userCoeffsCopy := &upb.UserCoefficients{Coeffs: coeffsCopy}
	score5 := r.Rank(userCoeffsCopy, videoCoeffs)
	expectEq(t, score4, score5)
}

func makeCoeffVectors(numUsers, numVideos int) ([]*upb.UserCoefficients, []*vpb.VideoCoefficients) {
	users := make([]*upb.UserCoefficients, numUsers)
	for i := 0; i < numUsers; i++ {
		users[i] = &upb.UserCoefficients{Coeffs: makeRandomCoeffs()}
	}
	videos := make([]*vpb.VideoCoefficients, numVideos)
	for i := 0; i < numVideos; i++ {
		videos[i] = &vpb.VideoCoefficients{Coeffs: makeRandomCoeffs()}
	}
	return users, videos
}

func testScoreDistribution(t *testing.T, r Ranker) {
	rand.Seed(42)
	scores := make(map[uint64]struct{})
	numUsers := 5
	numVideos := 7
	users, videos := makeCoeffVectors(numUsers, numVideos)
	for u := 0; u < numUsers; u++ {
		for v := 0; v < numVideos; v++ {
			s := r.Rank(users[u], videos[v])
			scores[s] = struct{}{}
		}
	}
	t.Logf("%d unique scores out of %d", len(scores), numUsers*numVideos)
	if float64(len(scores)) < 0.9*float64(numUsers*numVideos) {
		t.Errorf("insufficient score distribution for a ranker, %d unique values out of %d",
			len(scores), numUsers*numVideos)
	}
}

func TestBcryptRanker(t *testing.T) {
	t.Run("bcrypt/simple", func(t *testing.T) {
		r := BcryptRanker{}
		testSimple(t, &r)
	})
	t.Run("bcrypt/dist", func(t *testing.T) {
		r := BcryptRanker{}
		testScoreDistribution(t, &r)
	})
}
