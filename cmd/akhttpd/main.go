package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tkw1536/akhttpd"
)

func main() {
	r, err := akhttpd.NewGitHubKeyRepository(akhttpd.GitHubKeyRepoOptions{
		GitHubToken:     token,
		UpstreamTimeout: apiTimeout,
		MaxCacheSize:    cacheBytes,
		MaxCacheAge:     cacheTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}

	h := &akhttpd.Handler{KeyRepository: r}
	h.RegisterFormatter("", akhttpd.AuthorizedKeysFormatter{})
	h.RegisterFormatter("sh", akhttpd.ShellKeysFormatter{})

	h.LoadDefaultFiles()
	if indexHTMLPath != "" {
		indexHTMLBytes, err := ioutil.ReadFile(indexHTMLPath)
		if err != nil {
			log.Fatal(err)
			return
		}
		h.IndexHTML = string(indexHTMLBytes)
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
