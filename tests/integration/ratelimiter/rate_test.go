package ratelimiter_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yerassyldanay/requestmaker/connections/redisconn"
	"github.com/yerassyldanay/requestmaker/pkg/configx"
	"github.com/yerassyldanay/requestmaker/provider/ratelimiter"
	"go.uber.org/zap"
)

func getRateLimiterWithConnection(t *testing.T) *ratelimiter.ExternalReqLimit {
	logger, err := zap.NewDevelopment()
	require.NoErrorf(t, err, "failed to create dev logger")

	conf, err := configx.NewConfiguration()
	require.NoError(t, err)

	client, err := redisconn.NewRedisConnection(conf.RedisHost, conf.RedisPort, conf.RedisTestDatabase)
	require.NoError(t, err)

	statusCmd := client.FlushDB(context.Background())
	require.NoErrorf(t, statusCmd.Err(), "failed to clear database to run tests")

	return ratelimiter.NewExternalReqLimit(logger, client)
}

func TestRateLimiter(t *testing.T) {
	rateLimiterConn := getRateLimiterWithConnection(t)

	urlFound, err := url.Parse("https://google.com")
	require.NoErrorf(t, err, "failed to parse url")

	urlNotFound, err := url.Parse("https://notfound.com")
	require.NoErrorf(t, err, "failed to parse url")

	{
		err := rateLimiterConn.Record(context.Background(), *urlFound)
		require.NoErrorf(t, err, "failed to write a record to in-memory db")

		allowed, err := rateLimiterConn.Allowed(context.Background(), *urlFound)
		require.NoErrorf(t, err, "failed to check element in in-memory db")
		require.Falsef(t, allowed, "url should not be allowed as request was made recently")

		allowed, err = rateLimiterConn.Allowed(context.Background(), *urlNotFound)
		require.NoErrorf(t, err, "failed to check element in in-memory db")
		require.Truef(t, allowed, "url (not in in-memory db) is expected to be allowed")
	}
}
