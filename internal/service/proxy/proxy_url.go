package proxy

import (
	"net/url"

	"github.com/khatibomar/kaedama/internal/domain/dto"
	domainError "github.com/khatibomar/kaedama/internal/domain/error"
)

func (s *Service) ProxyURL(requestURL string) (*dto.ProxyResult, error) {
	url, err := url.Parse(requestURL)
	if err != nil {
		return nil, domainError.NewErrValidation(err)
	}

	// TODO add url processing
	_ = url

	return nil, nil
}
