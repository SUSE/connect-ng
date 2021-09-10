package connect

// StringSet is used to implement a set of strings
type StringSet struct {
	m map[string]struct{}
}

// NewStringSet constructor for StringSet.
// The set can be initialized with zero or more strings.
func NewStringSet(strs ...string) StringSet {
	ss := StringSet{}
	ss.m = make(map[string]struct{})
	for _, s := range strs {
		ss.m[s] = struct{}{}
	}
	return ss
}

// Add string(s) to set
func (ss StringSet) Add(strs ...string) {
	for _, s := range strs {
		ss.m[s] = struct{}{}
	}
}

// Delete s from set
func (ss StringSet) Delete(s string) {
	delete(ss.m, s)
}

// Contains returns true if s is in the set
func (ss StringSet) Contains(s string) bool {
	_, ok := ss.m[s]
	return ok
}

// Len returns the number of strings in the set
func (ss StringSet) Len() int {
	return len(ss.m)
}
