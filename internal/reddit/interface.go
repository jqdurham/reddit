package reddit

import "context"

// Waiter provides synchronization of http requests.
//
//go:generate mockery --name Waiter
type Waiter interface {
	Wait(ctx context.Context) error
}

// ListingFetcher declares the client's ability to fetch listings.
//
//go:generate mockery --name ListingFetcher
type ListingFetcher interface {
	FetchListing(ctx context.Context, path string) (*Listing, error)
	FetchAllListings(ctx context.Context, path string) ([]*Listing, error)
}
