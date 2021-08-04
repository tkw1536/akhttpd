package akhttpd

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
)

// handlePathOrFallback sends filepath to w with the provided content type
//
// When filepath does not exist (or is the empty string), returns fallbackBytes.
// Otherwise returns an error.
func handlePathOrFallback(w http.ResponseWriter, filepath string, fallbackBytes []byte, contentType string) error {
	var bytes []byte
	var err error = os.ErrNotExist

	// read bytes and error from the provided file
	if filepath != "" {
		bytes, err = os.ReadFile(filepath)
	}

	// if we could not find the error, use fallback!
	if errors.Is(err, os.ErrNotExist) {
		bytes = fallbackBytes
		err = nil
	}

	// other unknown error occured; something went wrong!
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	// write
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(bytes)
	return err
}

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
