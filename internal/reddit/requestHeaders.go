package reddit

import (
	"net/http"
)

type headerOptFunc func(h http.Header)

func withBearer(token string) headerOptFunc {
	return func(h http.Header) {
		h.Set("Authorization", "Bearer "+token)
	}
}

func stdHeaders(opts ...headerOptFunc) http.Header {
	headers := http.Header{
		http.CanonicalHeaderKey("Content-Type"): []string{"application/json"},
		http.CanonicalHeaderKey("User-Agent"):   []string{"reddit-client/v0.0.1 (by /u/jqdurham)"},
	}
	for _, opt := range opts {
		opt(headers)
	}

	return headers
}
