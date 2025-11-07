package proxy

import (
	"net/url"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

type ValidationError struct {
	err error
}

func (e ValidationError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

type Result struct {
	ContentType string
	// TODO will have other fields
}

func (s *Service) URL(requestURL string) (*Result, error) {
	url, err := url.Parse(requestURL)
	if err != nil {
		return nil, ValidationError{err: err}
	}

	// TODO add url processing
	_ = url

	return &Result{}, nil
}
