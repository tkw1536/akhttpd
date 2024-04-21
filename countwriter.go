package akhttpd

import "io"

// spellchecker:words akhttpd

// CountWriter is an io.Writer that wraps an underlying writer.
// It is not safe for concurrent writing.
//
// It counts the the total number of bytes written.
// Furthermore, once a single write fails, all future writes are silently suppressed.
type CountWriter struct {
	Writer io.Writer

	count int
	err   error
}

// Write writes b into Writer.
func (w *CountWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return len(b), nil
	}

	n, err := w.Writer.Write(b)
	w.count += n
	w.err = err
	return n, err
}

// State returns the first error that occurred within the writer and the total number of bytes written up to that point.
func (w CountWriter) State() (int, error) {
	return w.StateWith(nil)
}

// StateWith returns the first error that occurred within the writer and the total number of bytes written up to that point.
// When err is not nil, returns the provided error instead of the internal error.
func (w CountWriter) StateWith(err error) (int, error) {
	if err == nil {
		err = w.err
	}
	return w.count, err
}
