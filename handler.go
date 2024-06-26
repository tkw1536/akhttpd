package akhttpd

import (
	"context"
	_ "embed" // include default robots.txt and index.html
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/tkw1536/akhttpd/pkg/format"
	"github.com/tkw1536/akhttpd/pkg/repo"
)

// spellchecker:words akhttpd

// Handler is the main akhttpd Server Handler.
// It implements http.Handler, see the ServerHTTP method.
type Handler struct {
	repo.KeyRepository
	Formatters map[string]format.Formatter

	SuffixHTMLPath string // if non-empty, path to append to every html response
	IndexHTMLPath  string // if non-empty, path to serve index.html from
	RobotsTXTPath  string // if non-empty, path to serve robots.txt from
}

// RegisterFormatter registers formatter as the formatter for the provided extension.
// When extension is empty, registers it for the path without an extension.
func (h *Handler) RegisterFormatter(extension string, formatter format.Formatter) {
	if h.Formatters == nil {
		h.Formatters = make(map[string]format.Formatter)
	}
	h.Formatters[strings.ToLower(extension)] = formatter
}

// WriteSuffix writes the html suffix to w
func (h Handler) WriteSuffix(w io.Writer) error {
	if h.SuffixHTMLPath == "" {
		return nil
	}

	// open the file for reading
	f, err := os.Open(h.SuffixHTMLPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// write the suffix into the writer
	_, err = io.Copy(w, f)
	return err
}

var handlerPath = regexp.MustCompile(`^/[a-zA-Z\d-@]+((\.[a-zA-Z]+)|/[a-zA-Z_]+)?/?$`)

//go:embed resources/index.min.html
var defaultIndexHTML []byte

//go:embed resources/robots.txt
var defaultRobotsTXT []byte

// ServerHTTP serves the main akhttpd server.
// It only answers to GET requests, all other requests are answered with Method Not Allowed.
// Whenever something goes wrong, responds with "Internal Server Error" and logs the error.
//
// This method only responds successfully to a few URLS.
// All other URLs result in a HTTP 404 Response.
//
//	GET /
//	GET /index.html
//
// When IndexHTMLPath is not the empty string, sends back the file with Status HTTP 200.
// When IndexHTMLPath is empty, it sends back a default index.html file.
//
//	GET /${username}
//	GET /${username}.${formatter}, GET /${username}/${formatter}
//
// Fetches SSH Keys for the provided user and formats them with formatter.
// When formatter is omitted, uses the default formatter.
// If the formatter or user do not exist, returns HTTP 404.
//
//	GET /robots.txt
//
// When RobotsTXTPath is not the empty string, sends back the file with Status HTTP 200.
// When RobotsTXTPath is empty, it sends back a default robots.txt file.
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
		err := handlePathOrFallback(w, h.IndexHTMLPath, defaultIndexHTML, "text/html")
		if err != nil {
			return
		}
		h.WriteSuffix(w)
	case path == "/robots.txt":
		handlePathOrFallback(w, h.RobotsTXTPath, defaultRobotsTXT, "text/plain")
	case path == "/favicon.ico": // performance optimization as web browsers frequently request this
		http.NotFound(w, r)

	case handlerPath.MatchString(path): // the main route, where the bulk of handling takes place
		path = strings.Trim(path, "/")
		var ext string

		// handle both '.' and '/' as an index
		idx := strings.IndexRune(path, '.')
		if idx == -1 {
			idx = strings.IndexRune(path, '/')
		}

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

	source, keys, err := h.KeyRepository.GetKeys(context.Background(), username)
	if err != nil {
		if _, isNotFound := err.(repo.UserNotFoundError); isNotFound {
			http.NotFound(w, r)
			return
		}

		if _, isLegalUnavailable := err.(repo.UserNotAvailableError); isLegalUnavailable {
			http.Error(w, "Unavailable for legal reasons", http.StatusUnavailableForLegalReasons)
			return
		}

		log.Printf("%s: Internal Server Error: %s", r.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	n, err := formatter.WriteTo(username, source, keys, r, w)
	if n == 0 && err != nil {
		log.Printf("%s: Internal Server Error: %s", r.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
