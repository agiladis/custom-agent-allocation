package v1

import (
	"github.com/agiladis/custom-agent-allocation/internal/service"
	"github.com/agiladis/custom-agent-allocation/internal/webhook"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type WebhookHandler struct {
	publisher service.Publisher
}

func NewWebhookHandler(pub service.Publisher) *WebhookHandler {
	return &WebhookHandler{publisher: pub}
}

func (h *WebhookHandler) Receive(c *fiber.Ctx) error {
	var payload webhook.QiscusWebhookPayload

	if err := c.BodyParser(&payload); err != nil {
		log.Error().Err(err).Msg("failed to parse webhook payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	if payload.RoomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "room_id is required",
		})
	}

	if err := h.publisher.Publish(c.Context(), payload.RoomID); err != nil {
		log.Error().Err(err).Msg("failed to publish room_id to Redis")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not enqueue job",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
