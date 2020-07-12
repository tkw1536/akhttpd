package main

import (
	"context"
	"net/http"
	"time"

	"github.com/die-net/lrucache"
	"github.com/google/go-github/v29/github"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

var githubClient *github.Client

func initClient() {
	// we create an oauthTransport iff we have a token
	var oauthTransport http.RoundTripper
	if token != "" {
		// make a cache for the transport
		ctx := context.Background()

		// create a new oauth2 client
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		oauthTransport = oauth2.NewClient(ctx, ts).Transport
	}

	// make a new transport based on the oauth transport
	transport := &httpcache.Transport{
		Transport:           oauthTransport,
		Cache:               lrucache.New(flagCacheBytes, flagCacheTimeout),
		MarkCachedResponses: true,
	}

	// make a new http client with a cache
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(flagTimeout) * time.Second,
	}

	// initialize the github client
	githubClient = github.NewClient(client)
	if githubClient == nil {
		panic("Unable to initialize GitHub client")
	}
}
