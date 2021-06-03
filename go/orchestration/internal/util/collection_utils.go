package util

// DeepCopyStringMap creates a deep copy of the src map and returns it.
//
// If src is nil, an empty map is returned.
func DeepCopyStringMap(src map[string]string) map[string]string {
	dest := make(map[string]string, len(src))
	for key, value := range src {
		dest[key] = value
	}
	return dest
}
