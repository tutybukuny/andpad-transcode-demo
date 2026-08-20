package healthcheck

import (
	"github.com/gofiber/fiber/v3"

	"transcode-demo/internal/models/response"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HealthCheck(c fiber.Ctx) error {
	return response.NewResponse[any]().WithMessage("OK").JSON(c)
}
