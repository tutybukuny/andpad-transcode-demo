package response

import (
	"github.com/gofiber/fiber/v3"

	"transcode-demo/pkg/cerrors"
)

type Response[T any] struct {
	Status     string       `json:"status"`
	StatusCode int          `json:"-"`
	Code       cerrors.Code `json:"code"`
	Data       T            `json:"data,omitempty"`
	Message    string       `json:"message,omitempty"`
}

func NewResponse[T any]() *Response[T] {
	return &Response[T]{}
}

func (r *Response[T]) WithData(data T) *Response[T] {
	r.Data = data
	return r
}

func (r *Response[T]) WithMessage(message string) *Response[T] {
	r.Message = message
	return r
}

func (r *Response[T]) WithStatus(status string) *Response[T] {
	r.Status = status
	return r
}

func (r *Response[T]) WithCode(code cerrors.Code) *Response[T] {
	r.Code = code
	return r
}

func (r *Response[T]) JSON(c fiber.Ctx) error {
	if r.StatusCode == 0 {
		r.StatusCode = fiber.StatusOK
	}
	if r.Code == 0 {
		r.Code = cerrors.Code(r.StatusCode)
	}
	r.Status = r.Code.String()

	return c.Status(r.StatusCode).JSON(r)
}
