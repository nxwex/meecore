package node

import "context"

type Service struct {
	repository NodeRepository
}

func NewService(repository NodeRepository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Get(ctx context.Context, id int64) (*Node, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) GetAll(ctx context.Context) ([]Node, error) {
	return s.repository.GetAll(ctx)
}
