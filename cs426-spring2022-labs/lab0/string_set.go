package string_set

import (
	// "fmt"
	"sync"
)

type StringSet interface {
	// Add string s to the StringSet and return whether the string was inserted
	// in the set.
	Add(key string) bool

	// Return the number of unique strings in the set
	Count() int

	// Return all strings matching a regex `pattern` within a range `[begin,
	// end)` lexicographically (for Part C)
	PredRange(begin string, end string, pattern string) []string
}

type LockedStringSet struct {
	mu  sync.Mutex
	set map[string]bool
}

func MakeLockedStringSet() LockedStringSet {
	return LockedStringSet{set: make(map[string]bool)}
}

func (stringSet *LockedStringSet) Add(key string) bool {
	stringSet.mu.Lock()
	exists := stringSet.set[key]
	if exists {
		defer stringSet.mu.Unlock()
		return false
	} else {
		stringSet.set[key] = true
		defer stringSet.mu.Unlock()
		return true
	}
}

func (stringSet *LockedStringSet) Count() int {
	stringSet.mu.Lock()
	defer stringSet.mu.Unlock()
	return len(stringSet.set)
}

func (stringSet *LockedStringSet) PredRange(begin string, end string, pattern string) []string {
	return make([]string, 0)
}
