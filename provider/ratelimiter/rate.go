package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type ExternalReqLimit struct {
	logger              *zap.Logger
	client              redis.Cmdable
	waitBetweenRequests time.Duration
}

func NewExternalReqLimit(logger *zap.Logger, client redis.Cmdable) *ExternalReqLimit {
	return &ExternalReqLimit{
		logger:              logger,
		client:              client,
		waitBetweenRequests: 1 * time.Second,
	}
}

var _ ExternalReqLimiter = (*ExternalReqLimit)(nil)

func (reqLimit *ExternalReqLimit) Allowed(ctx context.Context, urlParam url.URL) (bool, error) {
	stringCmd := reqLimit.client.Get(ctx, urlParam.String())
	err := stringCmd.Err()
	switch {
	case err == nil:
		// pass
	case errors.Is(err, redis.Nil) || err.Error() == "redis: nil":
		reqLimit.logger.Debug("key not found", zap.String("url", urlParam.String()))
		return true, nil
	case err != nil:
		reqLimit.logger.Debug("failed to get element from in-memory db", zap.Error(err))
		return false, fmt.Errorf("failed to get last request time by url. err: %v", err)
	}

	millisecond, err := stringCmd.Int64()
	if err != nil {
		reqLimit.logger.Debug("failed to parse int from in-memory db", zap.Error(err))
		return false, fmt.Errorf("failed to parse int. err: %v", err)
	}

	wasBefore := time.UnixMilli(millisecond).Before(time.Now().Add(reqLimit.waitBetweenRequests * -1))
	if !wasBefore {
		reqLimit.logger.Debug("rate limit is hit by this request", zap.String("url", urlParam.String()))
	}
	return wasBefore, nil
}

func (reqLimit *ExternalReqLimit) Record(ctx context.Context, url url.URL) error {
	statusCmd := reqLimit.client.Set(ctx, url.String(), time.Now().UnixMilli(), 0)
	return statusCmd.Err()
}
