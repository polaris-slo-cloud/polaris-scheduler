package util

// Copies the value of *src into *dest if src is set and its value is less than *dest.
func SetIfLessThanInt32(dest *int32, src *int32) {
	if src != nil && *src < *dest {
		*dest = *src
	}
}

// Copies the value of *src into *dest if src is set and its value is less than *dest.
func SetIfLessThanInt64(dest *int64, src *int64) {
	if src != nil && *src < *dest {
		*dest = *src
	}
}

// Copies the value of *src into *dest if src is set and its value is greater than *dest.
func SetIfGreaterThanInt32(dest *int32, src *int32) {
	if src != nil && *src > *dest {
		*dest = *src
	}
}

// Copies the value of *src into *dest if src is set and its value is greater than *dest.
func SetIfGreaterThanInt64(dest *int64, src *int64) {
	if src != nil && *src > *dest {
		*dest = *src
	}
}
