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

type Mode string

// Add more eventually
const (
	ModeValues     Mode = "values"
	ModeList       Mode = "list"
	ModeRandom     Mode = "random"
	ModeStochastic Mode = "stochastic"
	ModeStatic     Mode = "static"
)

type Response struct {
	Body   string
	Resp   *http.Response
	Time   time.Duration
	Result string
	Error  error
}
