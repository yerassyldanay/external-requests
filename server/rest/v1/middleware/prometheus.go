package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yerassyldanay/requestmaker/provider/metricsprovider"
)

func HttpRequestStats(httpMetrics *metricsprovider.HttpMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		httpMetrics.TotalRequests.WithLabelValues().Inc()

		c.Next()

		httpMetrics.ResponseStatus.With(prometheus.Labels{
			"code": fmt.Sprint(c.Writer.Status()),
		}).Inc()
		httpMetrics.Duration.With(prometheus.Labels{
			"code":   fmt.Sprint(c.Writer.Status()),
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		}).Observe(float64(time.Since(startedAt).Milliseconds()))
	}
}

func PrometheusHandler(reg *prometheus.Registry) gin.HandlerFunc {
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		promHandler.ServeHTTP(c.Writer, c.Request)
	}
}
