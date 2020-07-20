package akhttpd

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/die-net/lrucache"
	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

// GHKeyRepositoryOptions represent options to a github public key repository
type GHKeyRepositoryOptions struct {
	// the token for the GitHub API (if any)
	GitHubToken string

	// Timeout after which to produce an error
	UpstreamTimeout time.Duration

	// Maximum size of the cache in bytes, 0 to disable
	MaxCacheSize int64

	// Maximum Age of the Cache in seconds, <= 0 to never expire
	MaxCacheAge int64
}

type ghKeyRepository struct {
	*github.Client
}

// NewGitHubKeyRepository instantiates a new PublicKeyRepository from github
func NewGitHubKeyRepository(opts GHKeyRepositoryOptions) (PublicKeyRepository, error) {
	var repo ghKeyRepository

	// we create an oauthTransport iff we have a token
	var oauthTransport http.RoundTripper
	if opts.GitHubToken != "" {
		// make a cache for the transport
		ctx := context.Background()

		// create a new oauth2 client
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: opts.GitHubToken},
		)
		oauthTransport = oauth2.NewClient(ctx, ts).Transport
	}

	// make a new transport based on the oauth transport
	transport := &httpcache.Transport{
		Transport:           oauthTransport,
		Cache:               lrucache.New(opts.MaxCacheSize, opts.MaxCacheSize),
		MarkCachedResponses: true,
	}

	// make a new http client with a cache
	client := &http.Client{
		Transport: transport,
		Timeout:   opts.UpstreamTimeout,
	}

	// initialize the github client
	repo.Client = github.NewClient(client)
	if repo.Client == nil {
		return nil, errors.New("Unable to initialize GitHub client")
	}

	return &repo, nil
}

// GetKey gets the key for a given repository
func (gr ghKeyRepository) GetKey(context context.Context, username string) ([]string, error) {
	// fetch all keys of error
	keys, res, err := gr.Users.ListKeys(context, username, &github.ListOptions{})
	if err != nil {
		err = errors.Wrap(err, "Unable to retrieve keys")
		return nil, err
	}

	// if the user wasn't found, we are in an error state
	if res.StatusCode == http.StatusNotFound {
		return nil, errUserNotFound
	}

	// key strings
	keyStrings := make([]string, len(keys))
	for index, key := range keys {
		if key == nil {
			err = errors.Errorf("Invalid key %d for user %s", index, username)
			return nil, err
		}

		keyStrings[index] = key.GetKey()
	}

	// and return the key strings themselves
	return keyStrings, nil
}

var errUserNotFound = errors.New("User does not exist")

// IsNotFound checks if an error is a not found error
func (gr ghKeyRepository) IsNotFound(err error) bool {
	return err == errUserNotFound
}
