package akhttpd

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// Handler represents an akhttp server handler
type Handler struct {
	KeyRepository
	Formatters map[string]Formatter

	IndexHTML string // string content for 'index.html'
	RobotsTXT string // string content for 'robots.txt'
}

// RegisterFormatter registers formatter as the formatter for the provided extension.
// When extension is empty, registers it for the path without an extension.
func (h *Handler) RegisterFormatter(extension string, formatter Formatter) {
	if h.Formatters == nil {
		h.Formatters = make(map[string]Formatter)
	}
	h.Formatters[strings.ToLower(extension)] = formatter
}

var handlerPath = regexp.MustCompile(`^/[a-zA-Z\d-]+(\.([a-zA-Z])+)?/?$`)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// ensure that only a GET is used, we don't support anything else
	// this includes just requesting HEAD.
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path

	switch {
	case r.Method != http.MethodGet:
	case path == "/", path == "":
		h.handleFile(w, h.IndexHTML, "text/html")
	case path == "/robots.txt":
		h.handleFile(w, h.RobotsTXT, "text/plain")

	case path == "/favicon.ico": // performance optimization as webbrowsers frequently request this
		http.NotFound(w, r)

	case handlerPath.MatchString(path): // the main route, where the bulk of handling takes place
		path = strings.Trim(path, "/")
		var ext string
		idx := strings.IndexRune(path, '.')
		if idx != -1 {
			ext = path[idx+1:]
			path = path[:idx]
		}
		h.serveAuthorizedKey(w, r, path, ext)

	default: // everything else isn't found
		http.NotFound(w, r)
	}
}

// serveAuthorizedKey serves an authorized_keys file for a given user
func (h Handler) serveAuthorizedKey(w http.ResponseWriter, r *http.Request, username, formatName string) {
	formatter, hasFormatter := h.Formatters[strings.ToLower(formatName)]
	if !hasFormatter {
		http.NotFound(w, r)
		return
	}

	keys, err := h.KeyRepository.GetKeys(context.Background(), username)
	if err != nil {
		if _, isNotFound := err.(UserNotFoundError); isNotFound {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	formatter.Format(username, keys, w)
}
