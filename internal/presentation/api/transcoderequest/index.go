package transcoderequest

import (
	"transcode-demo/internal/models/response"

	"github.com/gofiber/fiber/v3"
	z "go.uber.org/zap"

	"transcode-demo/internal/services/transcoderequestsvc"
	"transcode-demo/pkg/cerrors"
)

type Handler struct {
	l           *z.Logger
	transReqSvc *transcoderequestsvc.Service
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

	return response.NewResponse[CreateTranscodeRequestResp]().WithData(CreateTranscodeRequestResp{RequestID: requestID}).JSON(c)
}
