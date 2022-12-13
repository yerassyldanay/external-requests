package ratelimiter

import (
	"context"
	"net/url"
)

type ExternalReqLimiter interface {
	Allowed(ctx context.Context, url url.URL) (bool, error)
	Record(ctx context.Context, url url.URL) error
}
