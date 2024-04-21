// Command akhttpd is the authorized_keys http daemon.
// It implements a RESTFUL API which serves authorized_keys files for every GitHub user.
//
// # API
//
// This daemon exposes the following API.
//
//	GET /
//	GET /index.html
//
// Returns a human-readable index document.
//
//	GET /${username}
//
// When requested from common command line clients, behave like /${username}/authorized_keys.
// Else, behave like /${username}.html`.
//
//	GET /${username}/authorized_keys
//
// Returns an authorized_keys file for the provided username.
// When successful, returns HTTP 200 along with appropriate Content-Disposition and Content-Type Headers.
// When the user does not exist, returns HTTP 404.
// When something goes wrong, returns HTTP 500.
//
//	GET /${username}.html
//
// Returns a user-facing page to display keys for the provided username.
// When successful, returns HTTP 200.
// When the user does not exist, returns HTTP 404.
// When something goes wrong, returns HTTP 500.
//
//	GET /${username}.sh
//
// Returns a shell script that automatically fills the file '.ssh/authorized_keys' with the keys for the requested user.
// Any non-existent directories are created.
// Existing files are overwritten.
// This script intended to be piped into /bin/sh using a command like
//
//	curl http://localhost:8080/username.sh | /bin/sh
//
// When the user does not exist, returns HTTP 404.
// When something goes wrong, returns HTTP 500.
//
//	GET /robots.txt
//
// Returns a robots.txt file.
//
//	GET /_/
//
// Optionally serves a static folder for more information and detailed documentation of the current server.
//
//	GET /_/upload/
//
// Optionally serves an interface for user uploads.
//
// # Configuration
//
// akhttpd can be configured using an environment variable as well as command line arguments.
//
//	host:port
//
// By default akhttpd listens on localhost, port 8080 only.
// To change this, pass an argument of the form 'host:port' to the akhttpd command.
//
//	GITHUB_TOKEN=token, -token TOKEN
//
// akhttpd interacts with the GitHub API.
// By default, this interaction is unauthenticated.
// Instead a GitHub Personal Access Token can be used.
// It does not need access to any Scopes.
// It should be provided using either the GITHUB_TOKEN environment variable or the -token flag.
//
//	-api-timeout duration
//
// When interacting with the GitHub API, akhttpd uses a default timeout of 1s.
// After this timeout expires, any response is considered invalid and an HTTP 500 is returned to the client.
// Use this flag to change the default timeout.
//
//	-cache-age duration, -cache-size bytes
//
// To avoid unnecessary GitHub API requests, akhttpd caches responses.
// Responses are cached for 1h by default, with a maximum cache size of 25kb.
// Use these flags to change the defaults.
//
//	-akpath path
//
// Before querying the GitHub API for a users' public keys first check this path on the filesystem.
// If a file corresponding to a requested username exists, treat that file as an 'authorized_keys' file
// and return only keys stored in there.
//
//	-index filename
//
// A sensible default index.html file is served on the root directory.
// Use this flag to select a different file instead.
//
//	-serve path
//
// akhttpd can in addition to the standard routes serve a '_' route.
// Use this flag to configure a directory to be served from this path
//
//	-allow-uploads, -upload-auth USER:PASSWORD
//
// akhttpd can optionally allow users to upload their own keys temporarily.
// This endpoint can also be password protected with the given username and password.
// The upload-auth can also be provided with the UPLOAD_AUTH environment variable.
// Providing this variable automatically implies -allow-uploads.
//
//	LEGAL_BLOCK=user1,user2
//
// For legal reasons it might be necessary to block specific users from being served using this service.
// To block a specific user, use the LEGAL_BLOCK variable.
// It contains a comma-separated list of users to be blocked.
package main

// spellchecker:words akhttpd akpath

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tkw1536/akhttpd"
	"github.com/tkw1536/akhttpd/legal"
	"github.com/tkw1536/akhttpd/pkg/format"
	"github.com/tkw1536/akhttpd/pkg/repo"
)

func main() {
	repos := make(repo.Combo, 0, 3)

	// create a repository for uploadable uploadable
	var uploadable repo.UploadableKeys
	if allowUploads {
		repos = append(repos, &uploadable)
	}

	// create the files directory first
	if akFilesPath != "" {
		log.Printf("will check for public keys in %s", akFilesPath)
		disk := repo.Disk{FS: os.DirFS(akFilesPath)}
		repos = append(repos, disk)
	}

	// create a github key repo
	gr, err := repo.NewGitHubKeys(repo.GitHubKeysOptions{
		Token:        token,
		Timeout:      apiTimeout,
		MaxCacheSize: cacheBytes,
		MaxCacheAge:  cacheTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	repos = append(repos, gr)

	// blacklist provided users
	r := &repo.Blocklisted{
		Repository: repos,
		Blocked:    blocked,
	}

	// make a handler
	h := &akhttpd.Handler{KeyRepository: r}

	sh := format.ShellScript{}
	html := format.HTML{Suffix: h.WriteSuffix}
	authorized_keys := format.AuthorizedKeys{}
	magic := format.Magic{AuthorizedKeys: authorized_keys, HTML: html}

	h.RegisterFormatter("", magic)
	h.RegisterFormatter("authorized_keys", authorized_keys)
	h.RegisterFormatter("sh", sh)
	h.RegisterFormatter("html", html)

	h.IndexHTMLPath = indexHTMLPath
	if indexHTMLPath != "" {
		log.Printf("loaded '/' from %s", indexHTMLPath)
	}

	h.SuffixHTMLPath = suffixHTMLPath
	if suffixHTMLPath != "" {
		log.Printf("loaded html suffix from %s", suffixHTMLPath)
	}

	if underscorePath != "" {
		log.Printf("serving '/_/' from %s", underscorePath)
		http.Handle("/_/", h.ServeUnderscore(underscorePath))
	}
	http.Handle("/", h)

	if allowUploads {
		log.Printf("enabling user uploads")
		uploadable.Prefix = "uploaded-"
		uploadable.WriteSuffix = h.WriteSuffix
		if uploadAuth != "" {
			log.Printf("enabling protected user uploads")
			uploadable.AuthUser, uploadable.AuthPassword, _ = strings.Cut(uploadAuth, ":")
		}

		http.Handle("/_/upload/", &uploadable)
	}

	// bind and listen to the server
	log.Printf("Listening on %s\n", bindAddress)
	if err := http.ListenAndServe(bindAddress, nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

var bindAddress = "localhost:8080"

// flags
var token = os.Getenv("GITHUB_TOKEN")
var blocked = (func(blocked string) []string {
	if blocked == "" {
		return nil
	}
	return strings.Split(blocked, ",")
})(os.Getenv("LEGAL_BLOCK"))
var cacheBytes int64 = 25 * 1000
var cacheTimeout = 1 * time.Hour
var apiTimeout = 1 * time.Second

var indexHTMLPath = ""
var suffixHTMLPath = ""
var underscorePath = ""
var akFilesPath = ""

var uploadAuth = os.Getenv("UPLOAD_AUTH")
var allowUploads = len(uploadAuth) > 0

func init() {
	var legalFlag bool
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Print legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Println("This executable contains code from several different go packages. ")
			fmt.Println("Some of these packages require licensing information to be made available to the end user. ")
			fmt.Print(legal.Notices)
			os.Exit(0)
		}
	}()

	flag.StringVar(&token, "token", token, "token for github authentication (can also be set by 'GITHUB_TOKEN' variable). ")
	flag.Int64Var(&cacheBytes, "cache-size", cacheBytes, "maximum in-memory cache size in bytes")
	flag.DurationVar(&cacheTimeout, "cache-age", cacheTimeout, "maximum time after which cache entries should expire")
	flag.DurationVar(&apiTimeout, "api-timeout", apiTimeout, "timeout for github API connection")
	flag.StringVar(&indexHTMLPath, "index", indexHTMLPath, "optional path to '/' serve. Assumed to be of mime-type html. ")
	flag.StringVar(&suffixHTMLPath, "suffix", suffixHTMLPath, "optional path to append to all html responses. Assumed to be of mime-type html. ")
	flag.StringVar(&underscorePath, "serve", underscorePath, "optional path to '_' static directory to serve. ")
	flag.StringVar(&akFilesPath, "akpath", akFilesPath, "optional path to check for additional authorized keys files")
	flag.BoolVar(&allowUploads, "allow-uploads", allowUploads, "serve the '/_/upload/' path to allow users to temporarily upload their own keys")
	flag.StringVar(&uploadAuth, "upload-auth", uploadAuth, "Protect '/_/upload/' with a 'username:password' combination")

	flag.Parse()

	// read command line arguments
	if flag.NArg() != 1 {
		return
	}

	// read the bind address
	bindAddress = flag.Arg(0)
}
