package cerrors

import (
	"errors"
)

func ErrNotFound(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(NotFound, msg[0], err)
}

func ErrUnauthenticated(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(Unauthorized, msg[0], err)
}

func ErrPermissionDenied(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(Forbidden, msg[0], err)
}

func ErrInternal(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(InternalServerError, msg[0], err)
}

func ErrFailedPrecondition(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(PreconditionFailed, msg[0], err)
}

func ErrInvalidArgument(err error, msg ...string) *CError {
	if len(msg) == 0 {
		msg = []string{err.Error()}
	}
	return newError(BadRequest, msg[0], err)
}

func Error(code Code, message string, errs ...error) *CError {
	return newError(code, message, errs...)
}

func newError(code Code, message string, errs ...error) *CError {
	if message == "" {
		message = code.String()
	}

	var err error
	if len(errs) > 0 {
		err = errs[0]
	}

	var cError *CError
	if errors.As(err, &cError) && cError != nil {
		if cError.OriginalMessage == "" {
			cError.OriginalMessage = cError.Message
		}
		cError.Code = code
		cError.Message = message
		return cError
	}

	oriMsg := ""
	if err != nil {
		oriMsg = err.Error()
	}

	return &CError{
		Code:            code,
		Err:             err,
		Message:         message,
		OriginalMessage: oriMsg,
	}
}
