package v1

import (
	"context"

	"github.com/agiladis/custom-agent-allocation/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type ConfigHandler struct {
	svc service.ConfigService
}

func NewConfigHandler(svc service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

func (h *ConfigHandler) GetMaxLoad(c *fiber.Ctx) error {
	ctx := context.Background()
	val, err := h.svc.GetMaxLoad(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get max_load")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot read max_load",
		})
	}

	return c.JSON(fiber.Map{"value": val})
}

func (h *ConfigHandler) UpdateMaxLoad(c *fiber.Ctx) error {
	var body struct {
		Value int `json:"value"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}
	if body.Value < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "value cannot be smaller than 1",
		})
	}

	ctx := context.Background()
	if err := h.svc.SetMaxLoad(ctx, body.Value); err != nil {
		log.Error().Err(err).Msg("failed to update max_load")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot update max_load",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
