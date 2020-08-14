package akhttpd

import (
	"net/http"
	"os"
	"path/filepath"
)

func (Handler) handleFile(w http.ResponseWriter, text, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(text))
}

// LoadDefaultFiles loads the default content for index.html and robots.txt
func (h *Handler) LoadDefaultFiles() {
	h.IndexHTML = indexHTML
	h.RobotsTXT = robotsTxt
}

var indexHTML = `
<!doctype html>
<html lang="en">
<title>akhttpd - Authorized Keys HTTP Daemon</title>
<style>
a { color: blue; }
</style>
This domain serves an instance of <a href="https://github.com/tkw1536/akhttpd/" rel="noreferrer">akhttpd</a>. 
This page is not intended to be used by a web browser. 
Please consult the above link for more information. 
`

var robotsTxt = `
User-agent: *
Disallow: /
Allow: /$
Allow: /robots.txt$
`

// ServeUnderscore returns a handler for the '_' path at the given path
func (Handler) ServeUnderscore(path string) http.Handler {
	return http.StripPrefix("/_/", http.FileServer(noIndexFileSystem{http.Dir(path)}))
}

type noIndexFileSystem struct {
	http.FileSystem
}

func (fs noIndexFileSystem) Open(path string) (http.File, error) {
	file, err := fs.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		return nil, err
	}

	if stat.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := fs.FileSystem.Open(index); err != nil {
			file.Close()
			return nil, err
		}
	}

	return file, nil
}
