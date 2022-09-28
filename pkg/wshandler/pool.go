package wshandler

import (
	"net/http"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return new(Handler)
	},
}

// Handle instantiates a new handler, and calls it with the provided
func Handle(w http.ResponseWriter, r *http.Request, HandleFunc HandleFunc) {
	handler := pool.Get().(*Handler)
	defer pool.Put(handler)

	handler.Reset(HandleFunc)
	handler.ServeHTTP(w, r)
	handler.Reset(nil)
}
