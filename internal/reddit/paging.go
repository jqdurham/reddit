package reddit

import (
	"net/url"
	"strconv"
)

type Page struct {
	After, Before string
	Count         int
	Limit         int
}

func (p *Page) Values() url.Values {
	vals := url.Values{}

	if p.After != "" {
		vals.Set("after", p.After)
	}

	if p.Before != "" {
		vals.Set("before", p.Before)
	}

	if p.Count > 0 {
		vals.Set("count", strconv.Itoa(p.Count))
	}

	if p.Limit > 0 {
		vals.Set("limit", strconv.Itoa(p.Limit))
	}

	return vals
}
