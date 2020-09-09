package akhttpd

import "io"

// writeWrapper is an io.Writer that wraps an underlying writer.
// It is not safe for concurrent writing.
//
// It counts the the total number of bytes written.
// Furthermore, once a single read succeeds, all future writes are silently supressed.
type writeWrapper struct {
	w     io.Writer
	count int
	err   error
}

// Write writes b to the underlying writer.
func (w *writeWrapper) Write(b []byte) (int, error) {
	if w.err != nil {
		return len(b), nil
	}
	n, err := w.w.Write(b)
	w.count += n
	w.err = err
	return n, err
}

// State returns the total number of bytes written and any error that occured.
// If err is not nil, it will return e instead of the underlying error.
func (w *writeWrapper) State(err error) (int, error) {
	if err == nil {
		err = w.err
	}
	return w.count, err
}
