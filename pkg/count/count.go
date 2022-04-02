// Package count provides Writer
package count

import (
	"io"
	"sync"
)

// Writer is an io.Writer that wraps an underlying writer.
// It is not safe for concurrent writing.
//
// It counts the the total number of bytes written.
// Furthermore, once a single write fails, all future writes are silently supressed.
type Writer struct {
	Writer io.Writer

	count int
	err   error
}

// Reset resets this CountWriter before returning it to a pool
func (w *Writer) Reset() {
	w.Writer = nil
	w.count = 0
	w.err = nil
}

// Write writes b into Writer.
func (w *Writer) Write(b []byte) (int, error) {
	if w.err != nil {
		return len(b), nil
	}

	n, err := w.Writer.Write(b)
	w.count += n
	w.err = err
	return n, err
}

// State returns the first error that occured within the writer and the total number of bytes written up to that point.
func (w Writer) State() (int, error) {
	return w.StateWith(nil)
}

// StateWith returns the first error that occured within the writer and the total number of bytes written up to that point.
// When err is not nil, returns the provided error instead of the internal error.
func (w Writer) StateWith(err error) (int, error) {
	if err == nil {
		err = w.err
	}
	return w.count, err
}

// Pool is a pool of *Count.Writer objects
var pool = &sync.Pool{
	New: func() interface{} {
		return new(Writer)
	},
}

// Count runs f with a Writer using w.
// It returns cw.StateWith(err) where err is the return value of f.
func Count(w io.Writer, f func(cw *Writer) error) (int, error) {
	cw := pool.Get().(*Writer)
	cw.Writer = w

	defer cw.Reset()
	defer pool.Put(cw)

	return cw.StateWith(f(cw))
}
