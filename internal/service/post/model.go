package post

import "fmt"

// Post represents a topic.
type Post struct {
	Title string
	Ups   int
}

func (p *Post) String() string {
	return fmt.Sprintf("(%d) - %s \n", p.Ups, p.Title)
}

// AuthorPosts represents a count of posts created by a user.
type AuthorPosts struct {
	Author string
	Qty    int
}

func (a *AuthorPosts) String() string {
	return fmt.Sprintf("(%d) - %s \n", a.Qty, a.Author)
}
