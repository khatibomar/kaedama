package proxy

import (
	"net/url"

	domainError "github.com/khatibomar/kaedama/internal/domain/domainerror"
	"github.com/khatibomar/kaedama/internal/domain/dto"
)

func (s *Service) ProxyURL(requestURL string) (*dto.ProxyResult, error) {
	url, err := url.Parse(requestURL)
	if err != nil {
		return nil, domainError.NewErrValidation(err)
	}

	// TODO add url processing
	_ = url

	return &dto.ProxyResult{}, nil
}
