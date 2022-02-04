package string_set

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"testing"
)

func randomString(length int) string {
	chars := "abcdefghijklmnopqrstuv"

	out := ""
	for i := 0; i < length; i++ {
		randomChar := chars[rand.Int()%len(chars)]
		out = out + string(randomChar)
	}
	return out
}

func testSimpleAdds(t *testing.T, set StringSet) {
	if set.Count() != 0 {
		t.Errorf("Expected count to be 0")
	}
	if !set.Add("a") {
		t.Errorf("Expected insert of a to succeed")
	}
	if set.Add("a") {
		t.Errorf("Expected insert of a to fail, already inserted")
	}
	if set.Count() != 1 {
		t.Errorf("Count should be 1 after inserting a twice")
	}

	if !set.Add("b") {
		t.Errorf("Expected insert of b to succeed")
	}
	if set.Count() != 2 {
		t.Errorf("Expected count to update to 2")
	}
}

func testSimpleScans(t *testing.T, set StringSet) {
	set.Add("abc")
	set.Add("bbc")
	set.Add("def")
	set.Add("dbf")
	set.Add("123")
	set.Add("123b")

	result := set.PredRange("a", "c", ".*")
	sort.Strings(result)
	expectedResult := []string{"abc", "bbc"}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected abc and bbc to be between a and c")
	}

	result = set.PredRange("a", "z", "b")
	sort.Strings(result)
	expectedResult = []string{"abc", "bbc", "dbf"}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected scan to return strings with b")
	}
}

func testConcurrentAdds(t *testing.T, set StringSet) {
	ch := make(chan struct{})
	finalCh := make(chan struct{})

	numGoros := runtime.NumCPU()
	expectedNumItems := ((numGoros-1)/5 + 1) * 1000
	for i := 0; i < numGoros; i++ {
		go func(i int) {
			for x := 0; x < 100; x++ {
				for y := 0; y < 1000; y++ {
					set.Add(strconv.Itoa(i/5 + y*1000))
				}
			}
			ch <- struct{}{}
		}(i)
		// spinnup readers to continue reading count()
		go func(i int) {
			for {
				select {
				case <-finalCh:
					return
				default:
					if c := set.Count(); c > expectedNumItems {
						t.Errorf("Expected %d items, got %d", expectedNumItems, c)
					}
				}
			}
		}(i)
	}
	for i := 0; i < numGoros; i++ {
		<-ch
	}
	// stop readers
	close(finalCh)

	if c := set.Count(); c != expectedNumItems {
		t.Errorf("Expected %d items, got %d", expectedNumItems, c)
	}
}

func TestLocked(t *testing.T) {
	t.Run("locked/simple", func(t *testing.T) {
		set := MakeLockedStringSet()
		testSimpleAdds(t, &set)
	})
	t.Run("locked/concurrent", func(t *testing.T) {
		set := MakeLockedStringSet()
		testConcurrentAdds(t, &set)
	})
	t.Run("locked/scans", func(t *testing.T) {
		set := MakeLockedStringSet()
		testSimpleScans(t, &set)
	})
}

func TestStriped(t *testing.T) {
	for stripes := 1; stripes <= 20; stripes++ {
		t.Run(fmt.Sprintf("striped/simple/stripes=%d", stripes), func(t *testing.T) {
			set := MakeStripedStringSet(stripes)
			testSimpleAdds(t, &set)
		})
		t.Run(fmt.Sprintf("striped/concurrent/stripes=%d", stripes), func(t *testing.T) {
			set := MakeStripedStringSet(stripes)
			testConcurrentAdds(t, &set)
		})
		t.Run(fmt.Sprintf("striped/scans/stripes=%d", stripes), func(t *testing.T) {
			set := MakeStripedStringSet(stripes)
			testSimpleScans(t, &set)
		})
	}
}

func benchmarkAdds(b *testing.B, set StringSet) {
	b.RunParallel(func(pb *testing.PB) {
		i := rand.Intn(10)
		for j := 0; pb.Next(); j++ {
			set.Add(strconv.Itoa(i + j%100))
		}
	})
}

func benchmarkCounts(b *testing.B, set StringSet) {
	for i := 0; i < 10000; i++ {
		set.Add(randomString(10))
	}
	b.RunParallel(func(pb *testing.PB) {
		for j := 0; pb.Next(); j++ {
			if set.Count() == 0 {
				b.Fail()
			}
		}
	})
}

func benchmarkAddsAndCounts(b *testing.B, set StringSet) {
	b.RunParallel(func(pb *testing.PB) {
		i := rand.Intn(100)
		for j := 0; pb.Next(); j++ {
			set.Add(strconv.Itoa(i + j%1000))
			if set.Count() == 0 {
				b.Fail()
			}
		}
	})
}

func benchmarkScanSerial(b *testing.B, set StringSet) {
	for i := 0; i < 100000; i++ {
		set.Add(randomString(10))
	}
	for i := 0; i < b.N; i++ {
		set.PredRange("1", "8", "2.*4")
	}
}

func benchmarkScanParallel(b *testing.B, set StringSet) {
	for i := 0; i < 100000; i++ {
		set.Add(randomString(10))
	}
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			set.PredRange("1", "8", "2.*4")
		}
	})
}

func BenchmarkLockedStringSet(b *testing.B) {
	b.Run("adds", func(b *testing.B) {
		set := MakeLockedStringSet()
		benchmarkAdds(b, &set)
	})
	b.Run("counts", func(b *testing.B) {
		set := MakeLockedStringSet()
		benchmarkCounts(b, &set)
	})
	b.Run("adds+counts", func(b *testing.B) {
		set := MakeLockedStringSet()
		benchmarkAddsAndCounts(b, &set)
	})
	b.Run("scans:serial", func(b *testing.B) {
		set := MakeLockedStringSet()
		benchmarkScanSerial(b, &set)
	})
	b.Run("scans:parallel", func(b *testing.B) {
		set := MakeLockedStringSet()
		benchmarkScanParallel(b, &set)
	})
}

func BenchmarkStripedStringSet(b *testing.B) {
	stripeCounts := []int{1, 2, runtime.NumCPU(), runtime.NumCPU() * 2, runtime.NumCPU() * 8}
	for _, count := range stripeCounts {
		b.Run(fmt.Sprintf("adds/stripes=%d", count), func(b *testing.B) {
			set := MakeStripedStringSet(count)
			benchmarkAdds(b, &set)
		})
		b.Run(fmt.Sprintf("counts/stripes=%d", count), func(b *testing.B) {
			set := MakeStripedStringSet(count)
			benchmarkCounts(b, &set)
		})
		b.Run(fmt.Sprintf("adds+counts/stripes=%d", count), func(b *testing.B) {
			set := MakeStripedStringSet(count)
			benchmarkAddsAndCounts(b, &set)
		})
		b.Run(fmt.Sprintf("scans:serial/stripes=%d", count), func(b *testing.B) {
			set := MakeStripedStringSet(count)
			benchmarkScanSerial(b, &set)
		})
		b.Run(fmt.Sprintf("scans:parallel/stripes=%d", count), func(b *testing.B) {
			set := MakeStripedStringSet(count)
			benchmarkScanParallel(b, &set)
		})
	}
}
