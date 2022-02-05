package string_set

import (
	// "fmt"
	// "math/rand"
	"regexp"
	"sync"
)

type shard struct {
	items map[string]bool
	mu    sync.RWMutex
}

type StripedStringSet struct {
	sets    []*shard
	mu      sync.RWMutex
	cc      int
	stripes int
}

func MakeStripedStringSet(stripeCount int) StripedStringSet {

	m := make([]*shard, stripeCount)
	for i := 0; i < stripeCount; i++ {
		m[i] = &shard{items: make(map[string]bool)}
	}
	c := 0
	stripe_c := stripeCount

	return StripedStringSet{sets: m, cc: c, stripes: stripe_c}
}

func (stringSet *StripedStringSet) Add(key string) bool {

	ru := int([]rune(key)[0]) % stringSet.stripes

	stringSet.sets[ru].mu.Lock()
	defer stringSet.sets[ru].mu.Unlock()

	if stringSet.sets[ru].items[key] {
		return false
	}

	stringSet.sets[ru].items[key] = true

	stringSet.mu.Lock()
	stringSet.cc += 1
	stringSet.mu.Unlock()

	return true
}

func (stringSet *StripedStringSet) Count() int {

	// total := 0

	stringSet.mu.Lock()
	defer stringSet.mu.Unlock()
	return stringSet.cc
}

func (stringSet *StripedStringSet) PredRange(begin string, end string, pattern string) []string {
	var out []string
	var wg sync.WaitGroup
	ch := make(chan string, 1000)

	wg.Add(stringSet.stripes)
	for i := 0; i < stringSet.stripes; i++ {
		j := i
		go func() {
			defer wg.Done()
			stringSet.singlePredRange(begin, end, pattern, j, ch)
		}()
	}

	wg.Wait()
	for len(ch) > 0 {
		out = append(out, <-ch)
	}

	// fmt.Printf("%+q", out)
	// fmt.Println("------------")
	return out
}

func (stringSet *StripedStringSet) singlePredRange(begin string, end string, pattern string, stripe int, ch chan string) {
	re := regexp.MustCompile(pattern)

	for k := range stringSet.sets[stripe].items {
		if k >= begin && k < end && re.Match([]byte(k)) {
			ch <- k
		}
	}
}
