package ratelimiter

import (
	"context"
	"net/url"
)

// ExternalReqLimiter limits the number of requests to external services
// This is an abrtcution, the real implementation depends on the developer
type ExternalReqLimiter interface {
	// checking whether making a requests to the url is allowed or not
	Allowed(ctx context.Context, url url.URL) (bool, error)
	// recording a request to url
	Record(ctx context.Context, url url.URL) error
}
