package errors

import (
	stderrors "errors"
	"net/http"

	"pr-review/internal/entity"
	"pr-review/internal/http/dto"
	"pr-review/internal/logging"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	var domainErr *entity.DomainError
	if !stderrors.As(err, &domainErr) {
		logging.Printf("ERROR: [%s %s] Internal server error: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "internal server error",
			},
		})
		return
	}

	var statusCode int
	switch domainErr.Code {
	case entity.ErrorCodeTeamExists:
		statusCode = http.StatusBadRequest
	case entity.ErrorCodePRExists, entity.ErrorCodePRMerged, entity.ErrorCodeNotAssigned, entity.ErrorCodeNoCandidate:
		statusCode = http.StatusConflict
	case entity.ErrorCodeNotFound:
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
		logging.Printf("ERROR: [%s %s] Domain error: %s - %s", c.Request.Method, c.Request.URL.Path, domainErr.Code, domainErr.Message)
	}

	c.JSON(statusCode, dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    string(domainErr.Code),
			Message: domainErr.Message,
		},
	})
}
