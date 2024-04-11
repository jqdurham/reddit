package post

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/jqdurham/reddit/internal/logger"
	"github.com/jqdurham/reddit/internal/reddit"
)

type Service struct {
	client reddit.ListingFetcher
	writer io.Writer
}

// NewService instantiates a Post service responsible for updating and reporting statistics.
func NewService(client reddit.ListingFetcher, writer io.Writer) *Service {
	return &Service{client: client, writer: writer}
}

// UpdateTopPosts fetches and reports the top posters for the provided subreddit.
func (s *Service) UpdateTopPosts(ctx context.Context, subreddit string) error {
	var (
		logr  = logger.FromContext(ctx)
		start = time.Now()
		posts []*Post
		err   error
	)

	defer func() {
		logr.Debug("update top posts", "subreddit", subreddit, "dur", time.Since(start), "posts", len(posts))
	}()

	posts, err = s.fetchTopPosts(ctx, subreddit)
	if err != nil {
		return fmt.Errorf("fetch top posts: %v: %w", subreddit, err)
	}

	out := make([]fmt.Stringer, len(posts))
	for i, post := range posts {
		out[i] = post
	}

	err = s.write(fmt.Sprintf("Top Posts (%s)", subreddit), out)
	if err != nil {
		return fmt.Errorf("write: %v: %w", subreddit, err)
	}

	return nil
}

// UpdateTopNAuthors fetches all posts in a subreddit to determine the top N most active posters.
func (s *Service) UpdateTopNAuthors(ctx context.Context, subreddit string, num int) error {
	var (
		logr  = logger.FromContext(ctx)
		start = time.Now()
		err   error
	)

	defer func() {
		logr.Debug("update top n authors", "subreddit", subreddit, "dur", time.Since(start))
	}()

	counts, err := s.fetchTopAuthors(ctx, subreddit)
	if err != nil {
		return fmt.Errorf("fetch top authors: %v: %w", subreddit, err)
	}

	authors := make([]string, 0, len(counts))
	for author := range counts {
		authors = append(authors, author)
	}

	slices.SortStableFunc(authors, func(a, b string) int {
		return cmp.Compare(counts[b], counts[a])
	})

	authorPosts := make([]fmt.Stringer, 0)
	for i, author := range authors {
		authorPosts = append(authorPosts, &AuthorPosts{Author: author, Qty: counts[author]})

		if i > num {
			break
		}
	}

	top := authorPosts[:min(num, len(authorPosts))]

	err = s.write(fmt.Sprintf("Top %d Authors (%s)", num, subreddit), top)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (s *Service) fetchTopAuthors(ctx context.Context, subreddit string) (map[string]int, error) {
	listings, err := s.client.FetchAllListings(ctx, "/r/"+subreddit)
	if err != nil {
		return nil, fmt.Errorf("fetch all listings: %w", err)
	}

	counts := map[string]int{}

	for _, listing := range listings {
		for _, kid := range listing.Segment.Children {
			if _, ok := counts[kid.Post.Author]; !ok {
				counts[kid.Post.Author] = 0
			}

			counts[kid.Post.Author]++
		}
	}

	return counts, nil
}

func (s *Service) fetchTopPosts(ctx context.Context, subreddit string) ([]*Post, error) {
	listing, err := s.client.FetchListing(ctx, "/r/"+subreddit+"/top")
	if err != nil {
		return nil, fmt.Errorf("fetch post listing: %w", err)
	}

	posts := make([]*Post, 0, len(listing.Segment.Children))
	for _, kid := range listing.Segment.Children {
		posts = append(posts, &Post{
			Title: kid.Post.Title,
			Ups:   kid.Post.Ups,
		})
	}

	return posts, nil
}

// Write prints the results of an update request since we are not persisting statistics,
// nor the data gathered to formulate them.
func (s *Service) write(title string, msg []fmt.Stringer) error {
	buf := bytes.NewBufferString("\n")
	buf.WriteString(title + "\n")
	buf.WriteString(strings.Repeat("-", 80) + "\n")

	for _, post := range msg {
		buf.WriteString(post.String())
	}

	buf.WriteString("\n")

	_, err := s.writer.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error writing msg: %w", err)
	}

	return nil
}
