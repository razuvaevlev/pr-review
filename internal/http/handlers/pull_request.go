package handlers

import (
	"net/http"

	"pr-review/internal/http/dto"
	"pr-review/internal/http/errors"
	"pr-review/internal/logging"
	"pr-review/internal/service"

	"github.com/gin-gonic/gin"
)

type PullRequestHandler struct {
	prService *service.PullRequestService
}

func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
	}
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	var req dto.CreatePRRequest
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

	if err := req.Validate(); err != nil {
		logging.Printf("ERROR: [%s %s] Validation failed: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr, err := h.prService.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := dto.PullRequestResponse{
		PR: dto.FromEntity(pr),
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusCreated, response)
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var req dto.MergePRRequest
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

	if err := req.Validate(); err != nil {
		logging.Printf("ERROR: [%s %s] Validation failed: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr, err := h.prService.MergePR(req.PullRequestID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := dto.PullRequestResponse{
		PR: dto.FromEntity(pr),
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req dto.ReassignReviewerRequest
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

	if err := req.Validate(); err != nil {
		logging.Printf("ERROR: [%s %s] Validation failed: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := dto.ReassignResponse{
		PR:         dto.FromEntity(pr),
		ReplacedBy: replacedBy,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}
