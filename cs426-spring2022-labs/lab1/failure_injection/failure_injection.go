package failure_injection

import (
	"math/rand"
	"sync/atomic"
	"time"

	pb "cs426.yale.edu/lab1/failure_injection/proto"
)

var gInjectionConfig atomic.Value

func init() {
	ClearInjectionConfig()
	rand.Seed(time.Now().UnixNano())
}

func ClearInjectionConfig() {
	SetInjectionConfig(0, 0, 0)
}

func SetInjectionConfig(sleepNs, failureRate, responseOmissionRate int64) {
	SetInjectionConfigPb(&pb.InjectionConfig{
		SleepNs:              sleepNs,
		FailureRate:          failureRate,
		ResponseOmissionRate: responseOmissionRate,
	})
}

func SetInjectionConfigPb(config *pb.InjectionConfig) {
	gInjectionConfig.Store(config)
}

func GetInjectionConfig() *pb.InjectionConfig {
	config := gInjectionConfig.Load()
	return config.(*pb.InjectionConfig)
}

// returns true one in n times;
// if n == 0, always returns false
func oneIn(n int64) bool {
	if n == 0 {
		return false
	}
	return rand.Int63n(n) == 0
}

// Simulate error injection. Intended to be called by the backend service
// goroutine at the beginning of each request.  Based on the current injection
// config, sleep for SleepNs, perhaps hang forever to similate response
// omissions; returns whether the caller should error out this request.
func MaybeInject() (shouldError bool) {
	config := GetInjectionConfig()
	time.Sleep(time.Nanosecond * time.Duration(config.SleepNs))

	shouldError = oneIn(config.FailureRate)
	shouldOmitResponse := oneIn(config.ResponseOmissionRate)
	if shouldOmitResponse {
		// block forever
		select {}
	}
	return
}
