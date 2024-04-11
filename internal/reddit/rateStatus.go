package reddit

import (
	"fmt"
	"net/http"
	"strconv"
)

// RateStatus represents Reddit's reporting of their rate limiter.
type RateStatus struct {
	Remaining float32
	Reset     uint8
	Used      uint8
}

func rateStatus(header http.Header) (*RateStatus, error) {
	var (
		err    error
		remain float64
		reset  int
		used   int
	)

	if v, ok := header["X-Ratelimit-Remaining"]; ok && len(v) == 1 {
		remain, err = strconv.ParseFloat(v[0], 32)
		if err != nil {
			return nil, fmt.Errorf("parse ratelimit remaining: %w", err)
		}
	}

	if v, ok := header["X-Ratelimit-Reset"]; ok && len(v) == 1 {
		reset, err = strconv.Atoi(v[0])
		if err != nil {
			return nil, fmt.Errorf("parse ratelimit reset: %w", err)
		}
	}

	if v, ok := header["X-Ratelimit-Used"]; ok && len(v) == 1 {
		used, err = strconv.Atoi(v[0])
		if err != nil {
			return nil, fmt.Errorf("parse ratelimit used: %w", err)
		}
	}

	return &RateStatus{
		Remaining: float32(remain),
		Reset:     uint8(reset),
		Used:      uint8(used),
	}, nil
}
