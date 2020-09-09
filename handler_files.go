package akhttpd

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handlePathOrFallback(w http.ResponseWriter, filepath, fallbackContent, contentType string) {
	// create a reader from the filepath
	var reader io.ReadCloser
	var err error
	if filepath != "" {
		reader, err = os.OpenFile(filepath, os.O_RDONLY, os.ModePerm)
	}

	// if something went wrong, or we didn't create a reader in the first place
	// we should use the fallback reader instead.
	if reader == nil || err != nil {
		reader = ioutil.NopCloser(strings.NewReader(fallbackContent))
	}

	writeFile(w, reader, contentType)
}

func writeFile(w http.ResponseWriter, reader io.ReadCloser, contentType string) {
	bytes, err := ioutil.ReadAll(reader)
	defer reader.Close()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(bytes)
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
Allow: /_/
Allow: /robots.txt$
`

// ServeUnderscore returns an http.Handler that serves the provided filesystem path under the prefix '_'.
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
