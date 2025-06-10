package service

import (
	"context"
	"fmt"

	"github.com/agiladis/custom-agent-allocation/internal/repository"
)

type ConfigService interface {
	GetMaxLoad(ctx context.Context) (int, error)
	SetMaxLoad(ctx context.Context, newVal int) error
}

type configService struct {
	repo *repository.ConfigRepository
}

func NewConfigService(repo *repository.ConfigRepository) ConfigService {
	return &configService{repo: repo}
}

func (s *configService) GetMaxLoad(ctx context.Context) (int, error) {
	return s.repo.GetMaxLoad(ctx)
}

func (s *configService) SetMaxLoad(ctx context.Context, newVal int) error {
	if newVal < 1 {
		return fmt.Errorf("max_load cannot be smaller than 1")
	}

	return s.repo.UpdateMaxLoad(ctx, newVal)
}
