package models

import (
	"net/http"
	"time"
)

type Request struct {
	Name   string
	Path   string
	Method string
	Expect Expectation
	Body   []byte
}

type Expectation struct {
	StatusCode int
}

type Response struct {
	Body   string
	Resp   *http.Response
	Time   time.Duration
	Result string
	Error  error
}
