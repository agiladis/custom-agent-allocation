package qiscus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/agiladis/custom-agent-allocation/internal/config"
)

type Client struct {
	baseURL   string
	appID     string
	secretKey string
	http      *http.Client
}

type agentWrapper struct {
	Data struct {
		Agent struct {
			ID    int `json:"id"`
			Count int `json:"count"`
		} `json:"agent"`
	} `json:"data"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:   cfg.QiscusBaseURL,
		appID:     cfg.QiscusAppID,
		secretKey: cfg.QiscusSecretKey,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// call qiscus API: least active agent
func (c *Client) GetLeastActiveAgent(ctx context.Context) (agentID, count int, err error) {
	form := url.Values{}
	form.Set("source", "qiscus")

	url := fmt.Sprintf("%s/api/v1/admin/service/allocate_agent", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", c.appID)
	req.Header.Set("Qiscus-Secret-Key", c.secretKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("least-active returned %d: %s", resp.StatusCode, string(b))
	}

	var wrap agentWrapper
	if err := json.NewDecoder(resp.Body).Decode(&wrap); err != nil {
		return 0, 0, err
	}
	agent := wrap.Data.Agent
	return agent.ID, agent.Count, nil
}

// Assign agnet to room
func (c *Client) AssignAgent(ctx context.Context, roomID string, agentID int) error {
	form := url.Values{}
	form.Set("room_id", roomID)
	form.Set("agent_id", fmt.Sprintf("%d", agentID))

	url := fmt.Sprintf("%s/api/v1/admin/service/assign_agent", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Qiscus-App-Id", c.appID)
	req.Header.Set("Qiscus-Secret-Key", c.secretKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("assign returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
