package akhttpd

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/die-net/lrucache"
	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
)

// GitHubKeyRepo is an object that allows fetching ssh keys for GitHub Users using the GitHub API.
// It implements KeyRepository.
//
// The zero value is not ready to use, the caller should instantiate a GitHubClient first.
// See also NewGitHubKeyRepo.
type GitHubKeyRepo struct {
	*github.Client
}

// GitHubKeyRepoOptions represent options for a GitHubKeyRepo.
type GitHubKeyRepoOptions struct {
	// Token for GitHub Authentication.
	// Leave blank for anonymous requests; these might be subject to rate limiting.
	Token string

	// Timeout is the Timeout for requests to GitHub.
	// The zero value indiciates no timeout.
	Timeout time.Duration

	// MaxCacheSize is the maximum size of an internally used cache in bytes.
	// Leave blank to disable.
	MaxCacheSize int64

	// MaxCacheAge is the maximum age for any value in the cache.
	// Leave blank to never expire cache entires.
	MaxCacheAge time.Duration
}

var errClientReturnedNil = errors.New("github.NewClient returned nil")

// NewGitHubKeyRepo is a convenience method that instantiates NewGitHubKeyRepo.
// It reads options from opts, and returns a new GitHubKeyRepo.
func NewGitHubKeyRepo(opts GitHubKeyRepoOptions) (*GitHubKeyRepo, error) {
	var repo GitHubKeyRepo

	// using a token requires use of a transport.
	// we create one using oauth2.NewClient().
	var oauthTransport http.RoundTripper
	if opts.Token != "" {
		oauthTransport = oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: opts.Token},
			),
		).Transport
	}

	// create a new (cached) transport
	// based on the client above
	transport := &httpcache.Transport{
		Transport: oauthTransport,
		Cache: lrucache.New(
			opts.MaxCacheSize,
			int64(opts.MaxCacheAge.Seconds()),
		),
		MarkCachedResponses: true,
	}

	// finally make an http client with that cache
	// and the timeout above.
	client := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
	}

	// initialize the client
	repo.Client = github.NewClient(client)
	if repo.Client == nil {
		return nil, errClientReturnedNil
	}

	return &repo, nil
}

var errUserDoesNotExist = UserNotFoundError{errors.New("User does not exist")}

// GetKeys fetches keys from GitHub for the provided username.
// May internally cache results, as configured in the github.Client.
//
// If this function determines that a user does not exist, returns UserNotFoundError.
func (gr GitHubKeyRepo) GetKeys(context context.Context, username string) ([]ssh.PublicKey, error) {

	// this function works in two steps
	// - fetch the keys via the github api
	// - parse all the keys into ssh.PublicKey

	keys, res, err := gr.Users.ListKeys(context, username, &github.ListOptions{})
	if res != nil && res.StatusCode == http.StatusNotFound {
		return nil, errUserDoesNotExist
	}
	if err != nil {
		err = errors.Wrap(err, "Users.ListKeys failed")
		return nil, err
	}

	// Process all the keys in parallel.
	// Return the first error (if any)

	var wg sync.WaitGroup
	wg.Add(len(keys))

	pks := make([]ssh.PublicKey, len(keys))
	errChan := make(chan error, 1)

	for index := range keys {
		go parseKey(index, keys, pks, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	return pks, <-errChan // receive will not block because errChan is closed
}

// parseKey parses a single GitHub key and writes the result into pks.
// if an error occurs, it tries to send it to the error channel
func parseKey(index int, keys []*github.Key, pks []ssh.PublicKey, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	var err error
	if pks[index], _, _, _, err = ssh.ParseAuthorizedKey([]byte(keys[index].GetKey())); err == nil {
		return
	}

	select {
	case errChan <- err:
	default:
	}
}
