package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jqdurham/reddit/internal/logger"
)

const (
	unauthenticatedHost = "www.reddit.com"
	authenticatedHost   = "oauth.reddit.com"
)

// Client provides a mechanism to interact with Reddit's API.
type Client struct {
	clientID, secret string
	token            string
	httpClient       *http.Client
	limiter          Waiter
}

// NewClient creates and prepares a Client for interactions with Reddit's API.
func NewClient(clientID, secret string, httpClient *http.Client, limiter Waiter) *Client {
	return &Client{
		clientID:   clientID,
		secret:     secret,
		httpClient: httpClient,
		limiter:    limiter,
	}
}

// Login exchanges the provided credentials for a bearer token to be used with protected APIs.
func (c *Client) Login(ctx context.Context, username, password string) error {
	uri := &url.URL{Scheme: "https", Host: unauthenticatedHost, Path: "/api/v1/access_token"}

	req, err := c.prepareLoginRequest(ctx, uri, username, password)
	if err != nil {
		return err
	}

	res, err := c.send(ctx, req)
	if err != nil {
		return fmt.Errorf("fetch token: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return NewUnexpectedStatusError(http.MethodPost, uri.String(), res.StatusCode)
	}

	type accessTokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		Type        string `json:"type"`
	}

	response := &accessTokenResponse{}

	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return fmt.Errorf("token response decoding: %w", err)
	}

	c.token = response.AccessToken

	return nil
}

// FetchListing interacts with APIs that return listings.
func (c *Client) FetchListing(ctx context.Context, path string) (*Listing, error) {
	out := &Listing{}
	if err := c.fetchListing(ctx, path, out, nil); err != nil {
		return nil, err
	}

	return out, nil
}

// FetchAllListings pages through listingJson APIs and returns a set of settings.
// Each page request honors the rate limiter so processes that rely on this method will get slower
// updates.
func (c *Client) FetchAllListings(ctx context.Context, path string) ([]*Listing, error) {
	var (
		logr = logger.FromContext(ctx)
		page = &Page{Limit: 1000}
		out  []*Listing
	)

	for {
		listing := &Listing{}
		if err := c.fetchListing(ctx, path, listing, page); err != nil {
			return nil, err
		}

		page.After = listing.Segment.After
		page.Count += len(listing.Segment.Children)
		out = append(out, listing)

		logr.Debug("fetched page of listings", "path", path, "page", len(out))

		if listing.Segment.After == "" {
			break
		}
	}

	return out, nil
}

func (c *Client) uninitialized() bool {
	return c.clientID == "" || c.secret == ""
}

func (c *Client) validateLoginInputs(username, password string) error {
	if c.uninitialized() {
		return NewNotInitializedError()
	}

	if username == "" {
		return NewMissingInputError("username")
	}

	if password == "" {
		return NewMissingInputError("password")
	}

	return nil
}

func (c *Client) prepareLoginRequest(ctx context.Context, uri *url.URL, username, password string) (*http.Request, error) {
	if err := c.validateLoginInputs(username, password); err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Add("username", username)
	v.Add("password", password)
	v.Add("grant_type", "password")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	req.Header = stdHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(url.QueryEscape(c.clientID), url.QueryEscape(c.secret))

	return req, nil
}

func (c *Client) authenticated() error {
	if c.token == "" {
		return NewNotAuthenticatedError()
	}

	return nil
}

func (c *Client) prepareAuthenticatedRequest(ctx context.Context, path string, page *Page) (*http.Request, error) {
	if err := c.authenticated(); err != nil {
		return nil, err
	}

	qs := url.Values{}
	if page != nil {
		qs = page.Values()
	}

	uri := &url.URL{Scheme: "https", Host: authenticatedHost, Path: path, RawQuery: qs.Encode()}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %s | %w", path, err)
	}

	req.Header = stdHeaders(withBearer(c.token))

	return req, nil
}

func (c *Client) send(ctx context.Context, r *http.Request) (*http.Response, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}
	}

	res, err := c.httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return res, nil
}

func (c *Client) fetchListing(ctx context.Context, path string, listing *Listing, page *Page) error {
	logr := logger.FromContext(ctx)
	req, err := c.prepareAuthenticatedRequest(ctx, path, page)
	if err != nil {
		return err
	}

	res, err := c.send(ctx, req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	rate, err := rateStatus(res.Header)
	if err != nil {
		return err
	}

	logr.Debug("rate status", "used", rate.Used, "remaining", rate.Remaining, "reset", rate.Reset)

	switch res.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(res.Body).Decode(listing); err != nil {
			return fmt.Errorf("user listingJson: %w", err)
		}

		return nil
	case http.StatusTooManyRequests:
		return NewRateLimitExceededError(time.Duration(rate.Reset) * time.Second)
	default:
		return NewUnexpectedStatusError(http.MethodGet, path, res.StatusCode)
	}
}
