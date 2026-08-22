package transcoderequest

import (
	"github.com/gofiber/fiber/v3"
	z "go.uber.org/zap"

	"transcode-demo/internal/models/response"
	"transcode-demo/internal/services"
	"transcode-demo/internal/services/transcoderequestsvc"
	"transcode-demo/pkg/cerrors"
)

type Handler struct {
	l           *z.Logger
	transReqSvc services.ITranscodeRequestService
}

func NewHandler(l *z.Logger, transReqSvc *transcoderequestsvc.Service) *Handler {
	return &Handler{
		l:           l,
		transReqSvc: transReqSvc,
	}
}

// CreateTranscodeRequest POST /transcode
func (h *Handler) CreateTranscodeRequest(c fiber.Ctx) error {
	req := new(CreateTranscodeRequestReq)
	if err := c.Bind().Body(req); err != nil {
		return cerrors.ErrInvalidArgument(err)
	}

	requestID, err := h.transReqSvc.CreateTranscodeRequest(c, req.VideoURL)
	if err != nil {
		return err
	}

	return response.NewResponse[CreateTranscodeRequestResp]().
		WithData(CreateTranscodeRequestResp{RequestID: requestID}).
		JSON(c)
}

// GetTranscodeRequest GET /transcode-request/:id
func (h *Handler) GetTranscodeRequest(c fiber.Ctx) error {
	req := new(GetTranscodeRequestReq)
	if err := c.Bind().URI(req); err != nil {
		return cerrors.ErrInvalidArgument(err)
	}

	transReq, err := h.transReqSvc.GetTranscodeRequest(c, req.ID)
	if err != nil {
		return err
	}

	return response.NewResponse[GetTranscodeRequestResp]().
		WithData(GetTranscodeRequestResp{TranscodeRequest: transReq}).
		JSON(c)
}

// CancelTranscodeRequest POST /transcode-request/:id/cancel
func (h *Handler) CancelTranscodeRequest(c fiber.Ctx) error {
	req := new(CancelTranscodeRequestReq)
	if err := c.Bind().URI(req); err != nil {
		return cerrors.ErrInvalidArgument(err)
	}

	err := h.transReqSvc.CancelTranscodeRequest(c, req.ID)
	if err != nil {
		return err
	}
	return response.NewResponse[struct{}]().JSON(c)
}
