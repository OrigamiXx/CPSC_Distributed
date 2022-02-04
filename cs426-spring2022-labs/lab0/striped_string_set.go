package string_set

type StripedStringSet struct {
}

func MakeStripedStringSet(stripeCount int) StripedStringSet {
	return StripedStringSet{}
}

func (stringSet *StripedStringSet) Add(key string) bool {
	return false
}

func (stringSet *StripedStringSet) Count() int {
	return 0
}

func (stringSet *StripedStringSet) PredRange(begin string, end string, pattern string) []string {
	return make([]string, 0)
}
