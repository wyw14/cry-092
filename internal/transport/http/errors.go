package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-092/internal/domain/shared"
	"github.com/wyw14/cry-092/internal/middleware"
)

type errorResponse struct {
	Code        string              `json:"code"`
	Message     string              `json:"message"`
	FieldErrors []shared.FieldError `json:"field_errors"`
	RequestID   string              `json:"request_id"`
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code, message := "INTERNAL_ERROR", "服务暂时不可用"
	fields := []shared.FieldError{}
	var domainErr *shared.DomainError
	var inputErr *validationError
	if errors.As(err, &domainErr) {
		code, message, fields = domainErr.Code, domainErr.Message, domainErr.Fields
	}
	if errors.As(err, &inputErr) {
		status = http.StatusBadRequest
		code = "VALIDATION_FAILED"
		message = inputErr.message
		fields = []shared.FieldError{{Field: inputErr.field, Message: inputErr.message}}
	}
	switch {
	case errors.Is(err, shared.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, shared.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, shared.ErrConflict), errors.Is(err, shared.ErrInvalidState), errors.Is(err, shared.ErrVersionConflict):
		status = http.StatusConflict
	case domainErr != nil && domainErr.Cause == nil:
		status = http.StatusBadRequest
	}
	c.AbortWithStatusJSON(status, errorResponse{Code: code, Message: message, FieldErrors: fields, RequestID: middleware.CurrentRequestID(c)})
}
