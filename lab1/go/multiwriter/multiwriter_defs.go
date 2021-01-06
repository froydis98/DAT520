package multiwriter

import "dat520/lab1/go/errors"

// DO NOT EDIT THIS FILE.

type failureWriter int

// Write returns n = f, and never fails.
// This is used by the test functions.
func (f failureWriter) Write(p []byte) (n int, err error) {
	return int(f), nil
}

// errorsComparer returns true if the two input error slices are equal.
// This is used by the test functions.
func errorsComparer(a, b errors.Errors) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}