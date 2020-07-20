package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tkw1536/akhttpd"
)

func main() {
	r, err := akhttpd.NewGitHubKeyRepository(akhttpd.GHKeyRepositoryOptions{
		GitHubToken:     token,
		UpstreamTimeout: time.Duration(flagTimeout) * time.Second,
		MaxCacheSize:    flagCacheBytes,
		MaxCacheAge:     flagCacheTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}

	// setup handler
	http.HandleFunc("/", akhttpd.NewHandler(r, akhttpd.HandlerOpts{}))

	// bind
	log.Printf("Listening on %s\n", bindAddress)
	http.ListenAndServe(bindAddress, nil)
}

// args contains the command line arguments
var args []string
var token string
var flagCacheBytes int64
var flagCacheTimeout int64
var flagTimeout int
var bindAddress string

func init() {
	flag.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "token for github authentication (can also be set by 'GITHUB_TOKEN' variable). ")
	flag.Int64Var(&flagCacheBytes, "cache-size", 25*1000, "maximum in-memory cache size in bytes")
	flag.Int64Var(&flagCacheTimeout, "cache-age", 60*60, "maximum number of seconds after which cache entries should expire")
	flag.IntVar(&flagTimeout, "timeout", 1, "timeout in seconds after which to expire akhttpd")
	flag.Parse()

	// read command line arguments
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Missing 'bindAdress'")
		os.Exit(1)
	}

	// read the bind address
	bindAddress = args[0]
}
