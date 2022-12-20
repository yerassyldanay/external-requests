package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yerassyldanay/requestmaker/provider/metricsprovider"
	"github.com/yerassyldanay/requestmaker/server/rest/v1/middleware"
)

func prometheusHandler(reg *prometheus.Registry) gin.HandlerFunc {
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		promHandler.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()

	registrer := prometheus.NewRegistry()
	httpMetrics := metricsprovider.GetHttpMetrics(registrer)

	r.Use(middleware.HttpRequestStats(httpMetrics))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.GET("/metrics", prometheusHandler(registrer))

	r.Run(":8900")
}
