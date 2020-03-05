package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v29/github"
)

var bindAddress string
var client *github.Client

func main() {
	// read command line arguments
	if len(os.Args) != 2 {
		fmt.Println("Usage: akhttpd <bindAdress>")
		os.Exit(1)
	}

	// read the bind address
	bindAddress = os.Args[1]

	// initialize the github client
	client = github.NewClient(nil)
	if client == nil {
		panic("Unable to initialize GitHub client")
	}

	// setup handler
	http.HandleFunc("/", handleAKHTTPAD)

	// bind
	fmt.Printf("Listening on %s\n", bindAddress)
	http.ListenAndServe(bindAddress, nil)
}

var pathUsername = regexp.MustCompile(`^/[a-zA-Z\d-]+/?$`)

func handleAKHTTPAD(w http.ResponseWriter, r *http.Request) {
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

	username := strings.Split(path, "/")[1]

	// fetch the user data
	keys, isNotFound, err := getUserKeys(username)
	if err != nil && isNotFound {
		http.NotFound(w, r)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Printf("Error handling %s: %s", r.URL.Path, err.Error())
		w.Write([]byte("Service Unavailable\n"))
		return
	}

	// write the result
	w.Header().Add("Content-Disposition", "attachment; filename=\"authorized_keys\"")
	fmt.Fprintf(w, "# authorized_keys for %s, generated %s\n", username, time.Now().String())
	fmt.Fprint(w, strings.Join(keys, "\n")+"\n")
	return
}

func getUserKeys(username string) ([]string, bool, error) {
	// fetch all keys of error
	opts := &github.ListOptions{}
	keys, res, err := client.Users.ListKeys(context.Background(), username, opts)
	if err != nil {
		return nil, false, err
	}

	// key strings
	keyStrings := make([]string, len(keys))
	for index, key := range keys {
		if key == nil {
			return nil, false, fmt.Errorf("Invalid key %d for user %s", index, username)
		}

		keyStrings[index] = key.GetKey()
	}

	// and return the key strings themselves
	return keyStrings, res.StatusCode == http.StatusNotFound, nil
}
