package ranker

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"cs426.yale.edu/lab1/det_bcrypt"
	upb "cs426.yale.edu/lab1/user_service/proto"
	vpb "cs426.yale.edu/lab1/video_service/proto"
)

type Ranker interface {
	// takes in the coefficients for a user and a video, returns the overall score
	Rank(*upb.UserCoefficients, *vpb.VideoCoefficients) uint64
}

type BcryptRanker struct{}

func costHash(id uint64) int {
	hasher := fnv.New32()
	hasher.Write([]byte(strconv.FormatUint(id, 10)))
	const min = det_bcrypt.MinCost
	return int(hasher.Sum32())%(6-min) + min
}

func (br *BcryptRanker) Rank(
	userCoeffs *upb.UserCoefficients,
	videoCoeffs *vpb.VideoCoefficients,
) uint64 {
	var dotProduct uint64 = 0
	for featureId, userCoeff := range userCoeffs.GetCoeffs() {
		videoCoeff, ok := videoCoeffs.GetCoeffs()[featureId]
		if ok {
			// overflow u64 intentional, as long as it is deterministic
			dotProduct = dotProduct + userCoeff*videoCoeff
		}
	}

	pw := []byte(fmt.Sprintf("%020d", dotProduct))
	bytes, _ := det_bcrypt.GenerateFromPassword(pw, costHash(dotProduct))
	hasher := fnv.New64()
	hasher.Write(bytes)
	return hasher.Sum64()
}
