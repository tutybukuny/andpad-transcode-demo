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

// HealthCheck godoc
// @Summary      health check point
// @Description  API for health check
// @Tags         healthcheck
// @Success      200  {object}  response.Response[any]
// @Failure      500  {object}  response.Response[any]
// @Router       /health [get]
func (h *Handler) HealthCheck(c fiber.Ctx) error {
	return response.NewResponse[any]().WithMessage("OK").JSON(c)
}
