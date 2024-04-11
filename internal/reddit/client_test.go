// Only includes happy-path testing due to time constraints. The important pieces are mockable so edge
// cases can be tested. When hand-crafting a client from scratch, I prefer to use real samples
// from the API instead of slimmed down versions you see here.
package reddit_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/jqdurham/reddit/internal/reddit"
	"github.com/jqdurham/reddit/internal/reddit/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testUsername = "unitTester"
	testPassword = "UnITTestEr"

	loginURL        = "https://www.reddit.com/api/v1/access_token"
	listingURL      = "https://oauth.reddit.com/unit-test"
	firstListingURL = listingURL + "?limit=1000"
	lastListingURL  = listingURL + "?after=ou812&count=1&limit=1000"

	tokenJSON        = `{"access_token": "123", "token_type": "bearer", "expires_in": 86400, "scope": "*"}`
	firstListingJSON = `{"data": {
    "after": "ou812",
    "children": [{
        "data": {
          "title": "Unit test title",
          "name": "Unit test name",
          "ups": 1337,
          "author": "John Doe"
        }
      }]}}`
	lastListingJSON = `{"data": {
    "after": "",
    "children": [{
        "data": {
          "title": "Unit test title2",
          "name": "Unit test name2",
          "ups": 17,
          "author": "Jane Doe"
        }
      }]}}`
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	type args struct {
		clientID   string
		secret     string
		httpClient *http.Client
		limiter    reddit.Waiter
	}
	tests := []struct {
		name string
		args args
		want *reddit.Client
	}{
		{
			name: "Creates Client that implements ListingFetcher",
			args: args{
				clientID:   "unitTester",
				secret:     "secret",
				httpClient: http.DefaultClient,
				limiter:    nil,
			},
			want: &reddit.Client{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := reddit.NewClient(tt.args.clientID, tt.args.secret, tt.args.httpClient,
				tt.args.limiter)
			require.Implements(t, (*reddit.ListingFetcher)(nil), got)
		})
	}
}

func TestClient_FetchAllListings(t *testing.T) {
	t.Parallel()

	type fields struct {
		clientID   string
		secret     string
		token      string
		httpClient *http.Client
		limiter    func(t *testing.T) reddit.Waiter
	}

	type args struct {
		ctx  context.Context
		path string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*reddit.Listing
		errMsg string
	}{
		{
			name: "Loops when after is not empty, uses paging params and returns multiple listings",
			fields: fields{
				clientID: "clientID",
				secret:   "secret",
				token:    "tokenJson",
				httpClient: &http.Client{
					Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
						res := "{}"
						status := http.StatusBadRequest

						switch r.URL.String() {
						case loginURL:
							status = http.StatusOK
							res = tokenJSON
						case firstListingURL:
							status = http.StatusOK
							res = firstListingJSON
						case lastListingURL:
							status = http.StatusOK
							res = lastListingJSON
						}

						return &http.Response{
							StatusCode: status,
							Body:       io.NopCloser(strings.NewReader(res)),
						}, nil
					}),
				},
				limiter: func(t *testing.T) reddit.Waiter {
					t.Helper()
					m := mocks.NewWaiter(t)
					m.On("Wait", mock.Anything).Return(nil)

					return m
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/unit-test",
			},
			want: []*reddit.Listing{
				makeListing(firstListingJSON),
				makeListing(lastListingJSON),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := reddit.NewClient(tt.fields.clientID, tt.fields.secret, tt.fields.httpClient,
				tt.fields.limiter(t))

			err := c.Login(tt.args.ctx, testUsername, testPassword)
			require.NoError(t, err)

			got, err := c.FetchAllListings(tt.args.ctx, tt.args.path)

			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)
				require.Nil(t, got)

				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestClient_FetchListing(t *testing.T) {
	t.Parallel()
	type fields struct {
		clientID   string
		secret     string
		token      string
		httpClient *http.Client
		limiter    func(t *testing.T) reddit.Waiter
	}
	type args struct {
		ctx  context.Context
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *reddit.Listing
		wantErr bool
	}{
		{
			name: "Writes listings successfully",
			fields: fields{
				clientID: "clientID",
				secret:   "secret",
				token:    "tokenJson",
				httpClient: &http.Client{
					Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
						res := "{}"
						status := http.StatusBadRequest

						switch r.URL.String() {
						case loginURL:
							status = http.StatusOK
							res = tokenJSON
						case listingURL:
							status = http.StatusOK
							res = firstListingJSON
						}

						return &http.Response{
							StatusCode: status,
							Body:       io.NopCloser(strings.NewReader(res)),
						}, nil
					}),
				},
				limiter: func(t *testing.T) reddit.Waiter {
					t.Helper()
					m := mocks.NewWaiter(t)
					m.On("Wait", mock.Anything).Return(nil)

					return m
				},
			},
			args: args{
				ctx:  context.Background(),
				path: "/unit-test",
			},
			want: makeListing(firstListingJSON),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := reddit.NewClient(tt.fields.clientID, tt.fields.secret, tt.fields.httpClient,
				tt.fields.limiter(t))

			err := c.Login(tt.args.ctx, testUsername, testPassword)
			require.NoError(t, err)

			got, err := c.FetchListing(tt.args.ctx, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchListing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchListing() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	t.Parallel()
	type fields struct {
		clientID   string
		secret     string
		httpClient *http.Client
		limiter    reddit.Waiter
	}
	type args struct {
		ctx      context.Context
		username string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		errMsg string
	}{
		{
			name: "Valid credentials",
			fields: fields{
				clientID: "123",
				secret:   "456",
				httpClient: &http.Client{
					Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
						res := "{}"
						status := http.StatusBadRequest
						if r.URL.String() == loginURL {
							status = http.StatusOK
							res = tokenJSON
						}

						return &http.Response{
							StatusCode: status,
							Body:       io.NopCloser(strings.NewReader(res)),
						}, nil
					}),
				},
			},
			args: args{
				ctx:      context.Background(),
				username: "unitTester",
				password: "unitTester",
			},
		},
		{
			name: "Invalid credentials",
			fields: fields{
				clientID: "123",
				secret:   "456",
				httpClient: &http.Client{
					Transport: RoundTripperFunc(func(_ *http.Request) (*http.Response, error) {
						status := http.StatusUnauthorized

						return &http.Response{
							StatusCode: status,
							Body:       nil,
						}, nil
					}),
				},
			},
			args: args{
				ctx:      context.Background(),
				username: "unitTester",
				password: "unitTester",
			},
			errMsg: `unexpected status code 401 (POST https://www.reddit.com/api/v1/access_token)`,
		},
		{
			name:   "Empty username",
			fields: fields{clientID: "123", secret: "456"},
			args: args{
				ctx:      context.Background(),
				username: "",
				password: "unitTester",
			},
			errMsg: `missing required input: username`,
		},
		{
			name:   "Empty password",
			fields: fields{clientID: "123", secret: "456"},
			args: args{
				ctx:      context.Background(),
				username: "unitTester",
				password: "",
			},
			errMsg: `missing required input: password`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := reddit.NewClient(tt.fields.clientID, tt.fields.secret, tt.fields.httpClient,
				tt.fields.limiter)
			err := c.Login(tt.args.ctx, tt.args.username, tt.args.password)
			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)

				return
			}
			require.NoError(t, err)
		})
	}
}

func makeListing(data string) *reddit.Listing {
	listing := &reddit.Listing{}
	_ = json.Unmarshal([]byte(data), listing)

	return listing
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
