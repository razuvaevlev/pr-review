package handlers

import (
	"net/http"
	"strings"

	"pr-review/internal/config"
	"pr-review/internal/entity"
	"pr-review/internal/http/dto"
	"pr-review/internal/http/errors"
	"pr-review/internal/logging"
	"pr-review/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req dto.SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Printf("ERROR: [%s %s] Invalid request body: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		logging.Printf("ERROR: [%s %s] Validation failed: user_id is required", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id is required",
			},
		})
		return
	}

	if len(req.UserID) > config.MaxStringLength {
		logging.Printf("ERROR: [%s %s] Validation failed: user_id exceeds max length", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id cannot exceed 255 characters",
			},
		})
		return
	}

	user, err := h.userService.SetIsActive(req.UserID, req.IsActive)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := dto.UserResponse{
		User: user,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		logging.Printf("ERROR: [%s %s] Missing user_id query parameter", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id query parameter is required",
			},
		})
		return
	}
	if len(userID) > config.MaxStringLength {
		logging.Printf("ERROR: [%s %s] user_id exceeds max length: %d", c.Request.Method, c.Request.URL.Path, len(userID))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "user_id cannot exceed 255 characters",
			},
		})
		return
	}

	prs, err := h.userService.GetReviewPRs(userID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	shortPRs := make([]*entity.PullRequestShort, len(prs))
	for i, pr := range prs {
		shortPRs[i] = pr.ToShort()
	}

	response := dto.GetReviewResponse{
		UserID:       userID,
		PullRequests: shortPRs,
	}

	c.JSON(http.StatusOK, response)
}
