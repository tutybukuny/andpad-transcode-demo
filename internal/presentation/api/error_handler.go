package api

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

	"transcode-demo/internal/models/response"
	"transcode-demo/pkg/cerrors"
)

func DefaultStatusMapping(code cerrors.Code) int {
	if code > 999 {
		c := code
		for c > 999 {
			c /= 10
		}
		return int(c)
	}
	if code >= 100 {
		return int(code)
	}
	return http.StatusInternalServerError
}

func ErrorHandler(env string, httpStatusMappingFunc func(code cerrors.Code) int) func(ctx fiber.Ctx, err error) error {
	mappingFunc := httpStatusMappingFunc
	if mappingFunc == nil {
		mappingFunc = DefaultStatusMapping
	}
	return func(ctx fiber.Ctx, err error) error {
		// Statuscode defaults to 500
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			errCode := cerrors.Code(e.Code)
			return ctx.Status(e.Code).JSON(response.Response[any]{
				Status:  errCode.String(),
				Code:    errCode,
				Message: errCode.String(),
			})
		}

		clientError, ok := err.(*cerrors.CError)
		if !ok {
			return ctx.Status(code).JSON(response.Response[any]{
				Status:  cerrors.InternalServerError.String(),
				Code:    cerrors.InternalServerError,
				Message: cerrors.InternalServerError.String(),
			})
		}

		code = mappingFunc(clientError.Code)
		return ctx.Status(code).JSON(response.Response[any]{
			Status:  clientError.Code.String(),
			Code:    clientError.Code,
			Message: clientError.Message,
		})
	}
}
