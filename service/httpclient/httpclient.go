package httpclient

import "net/http"

type RequestMaker interface {
	Do(*http.Request) (*http.Response, error)
}
