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

// GitHubKeyRepoOptions are options to use for the GitHubKeyRepository
type GitHubKeyRepoOptions struct {
	// the token for the GitHub API (if any)
	GitHubToken string

	// Timeout after which to produce an error
	UpstreamTimeout time.Duration

	// Maximum size of the cache in bytes
	// By default, cache is disabled.
	MaxCacheSize int64

	// Maximum Age of the Cache in seconds, <= 0 to never expire
	MaxCacheAge time.Duration
}

// GitHubKeyRepo implements GitHubKeyRepo by using the GitHub API
// The zero value is not ready to use, a user should create their own github.Client
// or call NewGitHubKeyRepository()
type GitHubKeyRepo struct {
	*github.Client
}

// NewGitHubKeyRepository instantiates a new PublicKeyRepository with the provided options.
// This method is for convenience only, no private fields within GitHubKeyRepo are used.
func NewGitHubKeyRepository(opts GitHubKeyRepoOptions) (*GitHubKeyRepo, error) {
	var repo GitHubKeyRepo

	// if we have an OAuth Token
	// create an appropriate RoundTripper that uses that token
	var oauthTransport http.RoundTripper
	if opts.GitHubToken != "" {
		oauthTransport = oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: opts.GitHubToken},
			),
		).Transport
	}

	// make a new transport based on the oauth transport
	transport := &httpcache.Transport{
		Transport: oauthTransport,
		Cache: lrucache.New(
			opts.MaxCacheSize,
			int64(opts.MaxCacheAge.Seconds()),
		),
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
		return nil, errors.New("github.NewClient returned nil")
	}

	return &repo, nil
}

// GetKeys fetches github keys for the provided username.
// Results may be cached according to the options of the underlying GitHub API client.
// They may furthermore be
func (gr GitHubKeyRepo) GetKeys(context context.Context, username string) ([]ssh.PublicKey, error) {

	// this function works in two steps
	// - fetch the keys via the github api
	// - parse all the keys into ssh.PublicKey

	keys, res, err := gr.Users.ListKeys(context, username, &github.ListOptions{})
	if res != nil && res.StatusCode == http.StatusNotFound {
		return nil, UserNotFoundError{errors.New("User does not exist")}
	}
	if err != nil {
		err = errors.Wrap(err, "Users.ListKeys failed")
		return nil, err
	}

	// this second step is slightly more complicated than the above one.
	// Here we parse each of the keys in parallel

	var wg sync.WaitGroup
	wg.Add(len(keys))

	pks := make([]ssh.PublicKey, len(keys))
	errChan := make(chan error, 1)

	for index := range keys {
		go parseKey(index, keys, pks, &wg, errChan)
	}

	// wait for all the parseKey() goroutines to finish, then close the error channel
	// At this point an error (if any) is in the errChan buffer.
	wg.Wait()
	close(errChan)

	// return the results and the error
	return pks, <-errChan
}

// parseKey parses a single key
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
