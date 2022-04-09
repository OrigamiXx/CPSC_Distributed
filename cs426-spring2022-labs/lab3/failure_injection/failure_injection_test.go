package failure_injection

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func expectEq(t *testing.T, lhs, rhs int64) {
	if lhs != rhs {
		_, file, line, _ := runtime.Caller( /* skip */ 1)
		t.Errorf("%v and %v are expected to be equal. (%s:%d)", lhs, rhs, file, line)
	}
}

func testSimple(t *testing.T) {
	config := GetInjectionConfig()
	expectEq(t, config.SleepNs, 0)
	expectEq(t, config.FailureRate, 0)
	expectEq(t, config.ResponseOmissionRate, 0)

	config.SleepNs = 100
	config.FailureRate = 2000
	config.ResponseOmissionRate = 50000
	SetInjectionConfigPb(config)
	newConfig := GetInjectionConfig()
	if !proto.Equal(config, newConfig) {
		t.Errorf("Setting injection config failed!")
	}

	ClearInjectionConfig()
	config = GetInjectionConfig()
	expectEq(t, config.SleepNs, 0)
	expectEq(t, config.FailureRate, 0)
	expectEq(t, config.ResponseOmissionRate, 0)
}

const vMin = 10
const vMax = 5000

func checkVal(t *testing.T, expected, val int64) {
	expectEq(t, expected, val)
	if val < vMin || val >= vMax {
		t.Errorf("unexpected value %d", val)
	}
}

func checkInjectionConfig(t *testing.T) {
	config := GetInjectionConfig()
	val := config.SleepNs
	checkVal(t, val, config.SleepNs)
	checkVal(t, val, config.FailureRate)
	checkVal(t, val, config.ResponseOmissionRate)
}

func testConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := vMin; i < vMax; i++ {
		wg.Add(1)
		go func(i int) {
			x := int64(i)
			SetInjectionConfig(x, x, x)
			checkInjectionConfig(t)
			wg.Done()
		}(i)
	}
	wg.Wait()
	checkInjectionConfig(t)
}

func testLatencyInjection(t *testing.T) {
	sleepNs := int64(1000000000)
	SetInjectionConfig(sleepNs, 0, 0)
	numWorkers := 200
	ch := make(chan bool, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			start := time.Now()
			shouldError := MaybeInject()
			now := time.Now()
			elapsedNs := now.Sub(start).Nanoseconds()
			if elapsedNs < int64(float64(sleepNs)*0.8) || elapsedNs > int64(float64(sleepNs)*1.2) {
				t.Errorf("Injected %v ns instead of %v ns", elapsedNs, sleepNs)
			}
			ch <- shouldError
		}()
	}

	for i := 0; i < numWorkers; i++ {
		e := <-ch
		if e {
			t.Errorf("No error injection expected!")
		}
	}
}

func testFailureInjection(t *testing.T, rate int64) {
	SetInjectionConfig(0, rate, 0)
	totalReqs := int64(10000)
	numWorkers := 200
	expectedNumErrors := int64(totalReqs / rate)
	numErrors := int64(0)
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for j := int64(0); j < int64(totalReqs/int64(numWorkers)); j++ {
				shouldError := MaybeInject()
				if shouldError {
					atomic.AddInt64(&numErrors, 1)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if numErrors < int64(float64(expectedNumErrors)*0.8) ||
		numErrors > int64(float64(expectedNumErrors)*1.2) {
		t.Errorf("Injected %v errors instead of %v", expectedNumErrors, numErrors)
	}
}

func TestGlobalInjectionConfig(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		testSimple(t)
	})
	t.Run("concurrent", func(t *testing.T) {
		testConcurrent(t)
	})
}

func TestInjection(t *testing.T) {
	t.Run("latency", func(t *testing.T) {
		testLatencyInjection(t)
	})
	for _, rate := range []int64{1, 3, 10} {
		t.Run(fmt.Sprintf("failure/rate=%d", rate), func(t *testing.T) {
			testFailureInjection(t, rate)
		})
	}
}
