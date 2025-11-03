package proxy

type service struct{}

func New() *service {
	return &service{}
}
