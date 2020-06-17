package main

import (
	"net/http"
	"time"

	"github.com/die-net/lrucache"
	"github.com/google/go-github/v29/github"
	"github.com/gregjones/httpcache"
)

var githubClient *github.Client

func init() {
	// make a new http client with a cache
	client := &http.Client{
		Transport: httpcache.NewTransport(lrucache.New(flagCacheBytes, flagCacheTimeout)),
		Timeout:   time.Duration(flagTimeout) * time.Second,
	}

	// initialize the github client
	githubClient = github.NewClient(client)
	if githubClient == nil {
		panic("Unable to initialize GitHub client")
	}
}
