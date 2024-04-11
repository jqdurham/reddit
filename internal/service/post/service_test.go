package post_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jqdurham/reddit/internal/reddit"
	"github.com/jqdurham/reddit/internal/reddit/mocks"
	"github.com/jqdurham/reddit/internal/service/post"
	"github.com/stretchr/testify/require"
)

var listingJSON = `{
  "data": {
    "after": "",
    "children": [
      {
        "data": {
          "title": "Unit test title",
          "name": "Unit test name",
          "ups": 99999,
          "author": "John Doe"
        }
      },
      {
        "data": {
          "title": "Greatest shortstop",
          "name": "The Wizard",
          "ups": 1111,
          "author": "Ozzie Smith"
        }
      },
      {
        "data": {
          "title": "Opening Day Backflips",
          "name": "Backflippin'",
          "ups": 11,
          "author": "Ozzie Smith"
        }
      }
    ]
  }
}`

var testListing = func() *reddit.Listing {
	listing := &reddit.Listing{}
	_ = json.Unmarshal([]byte(listingJSON), listing)

	return listing
}()

var errMockedFailure = errors.New("mocked failure")

func TestService_UpdateTopNAuthors(t *testing.T) {
	t.Parallel()
	type fields struct {
		client func(t *testing.T) reddit.ListingFetcher
	}
	type args struct {
		ctx       context.Context
		subreddit string
		num       int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		errMsg string
		output string
	}{
		{
			name: "Writes top N authors",
			fields: fields{
				client: func(t *testing.T) reddit.ListingFetcher {
					t.Helper()
					m := mocks.NewListingFetcher(t)
					m.On("FetchAllListings", context.Background(),
						"/r/cardinals").Return([]*reddit.Listing{testListing}, nil)

					return m
				},
			},
			args: args{
				ctx:       context.Background(),
				subreddit: "cardinals",
				num:       10,
			},
			output: "\n" +
				"Top 10 Authors (cardinals)\n" +
				"--------------------------------------------------------------------------------\n" +
				"(2) - Ozzie Smith \n" +
				"(1) - John Doe \n\n",
		},
		{
			name: "Handles fetcher error getting posts",
			fields: fields{
				client: func(t *testing.T) reddit.ListingFetcher {
					t.Helper()
					m := mocks.NewListingFetcher(t)
					m.On("FetchAllListings", context.Background(),
						"/r/cubs").Return(nil, errMockedFailure)

					return m
				},
			},
			args: args{
				ctx:       context.Background(),
				subreddit: "cubs",
			},
			errMsg: "fetch top authors: cubs: fetch all listings: mocked failure",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			s := post.NewService(tt.fields.client(t), buf)
			err := s.UpdateTopNAuthors(tt.args.ctx, tt.args.subreddit, tt.args.num)
			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)
				require.Empty(t, buf)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.output, buf.String())
		})
	}
}

func TestService_UpdateTopPosts(t *testing.T) {
	t.Parallel()
	type fields struct {
		client func(t *testing.T) reddit.ListingFetcher
	}
	type args struct {
		ctx       context.Context
		subreddit string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		errMsg string
		output string
	}{
		{
			name: "Writes top posts",
			fields: fields{
				client: func(t *testing.T) reddit.ListingFetcher {
					t.Helper()
					m := mocks.NewListingFetcher(t)
					m.On("FetchListing", context.Background(),
						"/r/cardinals/top").Return(testListing, nil)

					return m
				},
			},
			args: args{
				ctx:       context.Background(),
				subreddit: "cardinals",
			},
			output: "\n" +
				"Top Posts (cardinals)\n" +
				"--------------------------------------------------------------------------------\n" +
				"(99999) - Unit test title \n" +
				"(1111) - Greatest shortstop \n" +
				"(11) - Opening Day Backflips \n\n",
		},
		{
			name: "Handles fetcher error top posts",
			fields: fields{
				client: func(t *testing.T) reddit.ListingFetcher {
					t.Helper()
					m := mocks.NewListingFetcher(t)
					m.On("FetchListing", context.Background(),
						"/r/cubs/top").Return(nil, errMockedFailure)

					return m
				},
			},
			args: args{
				ctx:       context.Background(),
				subreddit: "cubs",
			},
			errMsg: "fetch top posts: cubs: fetch post listing: mocked failure",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			s := post.NewService(tt.fields.client(t), buf)
			err := s.UpdateTopPosts(tt.args.ctx, tt.args.subreddit)
			if tt.errMsg != "" {
				require.EqualError(t, err, tt.errMsg)
				require.Empty(t, buf)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.output, buf.String())
		})
	}
}
