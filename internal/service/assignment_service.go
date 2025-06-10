package service

import (
	"context"
	"fmt"

	"github.com/agiladis/custom-agent-allocation/internal/config"
	"github.com/agiladis/custom-agent-allocation/internal/qiscus"
	"github.com/agiladis/custom-agent-allocation/internal/repository"
	"github.com/redis/go-redis/v9"
)

type AssignService interface {
	AssignCustomer(ctx context.Context, roomID string) error
}

type assignService struct {
	cfg    *config.Config
	rdb    *redis.Client
	repo   *repository.ConfigRepository
	qiscus *qiscus.Client
}

func NewAssignService(
	cfg *config.Config,
	rdb *redis.Client,
	repo *repository.ConfigRepository,
	qc *qiscus.Client,
) AssignService {
	return &assignService{cfg: cfg, rdb: rdb, repo: repo, qiscus: qc}
}

func (s *assignService) AssignCustomer(ctx context.Context, roomID string) error {
	// get max_load from Redis
	maxLoad, err := s.repo.GetMaxLoad(ctx)
	if err != nil {
		return fmt.Errorf("get max_load: %w", err)
	}

	// trigger call least-active agent
	agentID, load, err := s.qiscus.GetLeastActiveAgent(ctx)
	if err != nil {
		return fmt.Errorf("get least-active agent: %w", err)
	}

	// validate current load
	if load > maxLoad {
		// does not meet criteria → backoff
		return fmt.Errorf("agent %d load %d exceeds max_load %d", agentID, load, maxLoad)
	}

	// assign
	if err := s.qiscus.AssignAgent(ctx, roomID, agentID); err != nil {
		return fmt.Errorf("assign to agent %d: %w", agentID, err)
	}

	return nil
}
