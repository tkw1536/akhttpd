package akhttpd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// HandlerOpts represents the options of a handler
type HandlerOpts struct {
}

var pathUsername = regexp.MustCompile(`^/[a-zA-Z\d-]+/?$`)

// NewHandler makes a new handler for a public key repository
func NewHandler(repo PublicKeyRepository, opts HandlerOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// robots.txt
		if path == "/robots.txt" {
			fmt.Fprintf(w, "User-agent: *\nDisallow: /\n")
			return
		}

		// if not proper username
		if !pathUsername.MatchString(path) {
			http.NotFound(w, r)
			return
		}

		// fetch the key
		username := strings.Split(path, "/")[1]
		keys, err := repo.GetKey(context.Background(), username)

		// check for errors
		if err != nil {

			// not found
			if repo.IsNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found\n"))
				return
			}

			// general errors
			log.Printf(err.Error())
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable\n"))
			return
		}

		// write the result
		w.Header().Add("Content-Disposition", "attachment; filename=\"authorized_keys\"")
		fmt.Fprintf(w, "# authorized_keys for %s, generated %s\n", username, time.Now().String())
		fmt.Fprint(w, strings.Join(keys, "\n")+"\n")
		return
	}
}
