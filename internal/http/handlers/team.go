package handlers

import (
	"net/http"

	"pr-review/internal/config"
	"pr-review/internal/http/dto"
	"pr-review/internal/http/errors"
	"pr-review/internal/logging"
	"pr-review/internal/service"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) Add(c *gin.Context) {
	var team dto.TeamRequest
	if err := c.ShouldBindJSON(&team); err != nil {
		logging.Printf("ERROR: [%s %s] Invalid request body: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "invalid request body: " + err.Error(),
			},
		})
		return
	}

	if err := team.Validate(); err != nil {
		logging.Printf("ERROR: [%s %s] Validation failed: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	entityTeam := team.ToEntity()

	err := h.teamService.AddTeam(entityTeam)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	response := dto.TeamResponse{
		Team: entityTeam,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusCreated, response)
}

func (h *TeamHandler) Get(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		logging.Printf("ERROR: [%s %s] Missing team_name query parameter", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name query parameter is required",
			},
		})
		return
	}
	if len(teamName) > config.MaxStringLength {
		logging.Printf("ERROR: [%s %s] team_name exceeds max length: %d", c.Request.Method, c.Request.URL.Path, len(teamName))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "team_name cannot exceed 255 characters",
			},
		})
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, team)
}
