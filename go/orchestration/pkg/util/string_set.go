package util

var (
	_ StringSet = (*stringSetImpl)(nil)
)

// Represents a set of strings.
type StringSet interface {
	// Adds the specified entry to the set.
	Add(entry string)

	// Returns true if the specified entry is in this set, otherwise false.
	Has(entry string) bool

	// Deletes the specified entry from the set, if it was present.
	Delete(entry string)

	// Returns a slice containing all the entries in the set.
	Entries() []string
}

// Default implementation of StringSet
type stringSetImpl struct {
	entries map[string]bool
}

// Creates a new StringSet.
func NewStringSet() StringSet {
	return &stringSetImpl{
		entries: make(map[string]bool),
	}
}

func (me *stringSetImpl) Add(entry string) {
	me.entries[entry] = true
}

func (me *stringSetImpl) Has(entry string) bool {
	_, ok := me.entries[entry]
	return ok
}

func (me *stringSetImpl) Delete(entry string) {
	delete(me.entries, entry)
}

func (me *stringSetImpl) Entries() []string {
	ret := make([]string, len(me.entries))
	i := 0
	for key := range me.entries {
		ret[i] = key
	}
	return ret
}
