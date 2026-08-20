package middleware

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v3"
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/internal/constant"
	dbconstant "transcode-demo/pkg/db/constant"
	"transcode-demo/pkg/utils"
)

var defaultStackTraceBufLen = 4 << 10

func getRequestID(ctx context.Context) string {
	requestId, ok := ctx.Value(constant.RequestIdContext).(string)
	if !ok {
		requestId = ""
	}
	return requestId
}

func CustomRecover(l *z.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				stack := make([]byte, defaultStackTraceBufLen)
				length := runtime.Stack(stack, true)
				reqId, _ := c.Context().Value(constant.RequestIdContext).(string)
				l.Error("panic recovered", z.String("request_id", reqId), z.Error(err), z.ByteString("stack", stack[:length]))
				c.Status(fiber.StatusInternalServerError)
			}
		}()
		return c.Next()
	}
}

func RequestIDContext() fiber.Handler {
	return func(c fiber.Ctx) error {
		rid := utils.GetString(c.Response().Header.Peek(fiber.HeaderXRequestID))
		ctx := context.WithValue(c.Context(), constant.RequestIdContext, rid)
		c.SetContext(ctx)
		return c.Next()
	}
}

func GormTransaction(l *z.Logger, gormDB *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		tx := gormDB.Begin()
		ctx := context.WithValue(c.Context(), dbconstant.CtxDBKey, tx)
		c.SetContext(ctx)
		requestID := getRequestID(ctx)
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					var ok bool
					err, ok = r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, defaultStackTraceBufLen)
					length := runtime.Stack(stack, true)
					l.Error(
						"panic recovered",
						z.String("request_id", requestID),
						z.Error(err),
						z.ByteString("stack", stack[:length]),
					)
				}
			}()
			return c.Next()
		}()
		if err != nil {
			tErr := tx.Rollback().Error
			if tErr != nil {
				l.Error("Failed to rollback transaction", z.String("request_id", requestID), z.Error(tErr))
			}
			return err
		}

		err = tx.Commit().Error
		if err != nil {
			l.Error("Failed to commit transaction", z.String("request_id", requestID), z.Error(err))
			return err
		}

		return nil
	}
}

func Logging(l *z.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := getRequestID(c.Context())
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		if err != nil {
			l.Error(
				"failed handling request",
				z.String("request_id", requestID),
				z.String("method", c.Method()),
				z.String("path", c.Path()),
				z.Error(err),
				z.Duration("duration", duration),
			)
			return err
		}
		l.Info(
			"handle request successfully",
			z.String("request_id", requestID),
			z.String("method", c.Method()),
			z.String("path", c.Path()),
			z.Duration("duration", duration),
		)
		return nil
	}
}
