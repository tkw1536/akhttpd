// Command akhttpd is the authorized_keys http daemon.
// It implements a RESTFUL API which serves authorized_keys files for every GitHub user.
//
// API
//
// This daemon exposes the following API.
//
//  GET /
//  GET /index.html
// Returns a human-readable index document.
//
//  GET /${username}
//
// Returns an authorized_keys file for the provided username.
// When successful, returrns HTTP 200 along with appropriate Content-Disposition and Content-Type Headers.
// When the user does not exist, returns HTTP 404.
// When something goes wrong, returns HTTP 500.
//
//  GET /${username}.sh
//
// Returns a shell script that automatically fills the file '.ssh/authorized_keys' with the keys for the requested user.
// Any non-existent directories are created.
// Existing files are overwritten.
// This script intended to be piped into /bin/sh using a command like
//  curl http://localhost:8080/username.sh | /bin/sh
// When the user does not exist, returns HTTP 404.
// When something goes wrong, returns HTTP 500.
//
//  GET /robots.txt
//
// Returns a robots.txt file.
//
//  GET /_/
//
// Optionally serves a static folder for more information and detailed documentation of the current server.
//
// Configuration
//
// akhttpd can be configured using an environment variable as well as command line arguments.
//
//  host:port
//
// By default akhttpd listens on localhost, port 8080 only.
// To change this, pass an argument of the form 'host:port' to the akhttpd command.
//
//  GITHUB_TOKEN=token, -token TOKEN
//
// akhttpd interacts with the GitHub API.
// By default, this interaction is unauthenticated.
// Instead a GitHub Personal Access Token can be used.
// It does not need access to any Scopes.
// It should be provided using either the GITHUB_TOKEN environment variable or the -token flag.
//
//  -api-timeout duration
//
// When interacting with the GitHub API, akhttp uses a default timeout of 1s.
// After this timeout expires, any response is considered invalid and an HTTP 500 is returned to the client.
// Use this flag to change the default timeout.
//
//  -cache-age duration, -cache-size bytes
//
// To avoid unneccessary GitHub API requests, akhttpd caches responses.
// Respones are cached for 1h by default, with a maximum cache size of 25kb.
// Use these flags to change the defaults.
//
//  -index filename
//
// A sensible default index.html file is served on the root directory.
// Use this flag to select a different file instead.
//
//  -serve path
//
// akhttpd can in addition to the standard routes serve a '_' route.
// Use this flag to configure a directory to be served from this path.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tkw1536/akhttpd"
	"github.com/tkw1536/akhttpd/legal"
)

func main() {
	r, err := akhttpd.NewGitHubKeyRepo(akhttpd.GitHubKeyRepoOptions{
		Token:        token,
		Timeout:      apiTimeout,
		MaxCacheSize: cacheBytes,
		MaxCacheAge:  cacheTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}

	h := &akhttpd.Handler{KeyRepository: r}
	h.RegisterFormatter("", akhttpd.FormatterAuthorizedKeys{})
	h.RegisterFormatter("sh", akhttpd.FormatterShellScript{})

	h.IndexHTMLPath = indexHTMLPath
	if indexHTMLPath != "" {
		log.Printf("loaded '/' from %s", indexHTMLPath)
	}

	if underscorePath != "" {
		log.Printf("serving '/_/' from %s", underscorePath)
		http.Handle("/_/", h.ServeUnderscore(underscorePath))
	}
	http.Handle("/", h)

	// bind and listen to the server
	log.Printf("Listening on %s\n", bindAddress)
	if err := http.ListenAndServe(bindAddress, nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

var args []string
var bindAddress = "localhost:8080"

// flags
var token = os.Getenv("GITHUB_TOKEN")
var cacheBytes int64 = 25 * 1000
var cacheTimeout = 1 * time.Hour
var apiTimeout = 1 * time.Second

var indexHTMLPath = ""
var underscorePath = ""

func init() {
	var legalFlag bool
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Print legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Println("This executable contains code from several different go packages. ")
			fmt.Println("Some of these packages require licensing information to be made available to the end user. ")
			fmt.Println(legal.Notices)
			os.Exit(0)
		}
	}()

	flag.StringVar(&token, "token", token, "token for github authentication (can also be set by 'GITHUB_TOKEN' variable). ")
	flag.Int64Var(&cacheBytes, "cache-size", cacheBytes, "maximum in-memory cache size in bytes")
	flag.DurationVar(&cacheTimeout, "cache-age", cacheTimeout, "maximum time after which cache entries should expire")
	flag.DurationVar(&apiTimeout, "api-timeout", apiTimeout, "timeout for github API connection")
	flag.StringVar(&indexHTMLPath, "index", indexHTMLPath, "optional path to '/' serve. Assumed to be of mime-type html. ")
	flag.StringVar(&underscorePath, "serve", underscorePath, "optional path to '_' static directory to serve. ")
	flag.Parse()

	// read command line arguments
	if flag.NArg() != 1 {
		return
	}

	// read the bind address
	bindAddress = flag.Arg(0)
}
