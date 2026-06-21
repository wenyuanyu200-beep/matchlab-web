package questionnaire

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/auth"
)

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{repository: repository}
}

type submitRequest struct {
	Mode    string  `json:"mode"`
	Answers Answers `json:"answers"`
}

func (h *Handler) Submit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	var input submitRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
	if input.Mode != "activity" {
		writeError(c, http.StatusBadRequest, "invalid_request", "mode must be activity")
		return
	}
	input.Answers = normalizeAnswers(input.Answers)
	generated := GenerateProfile(input.Mode, input.Answers)
	questionnaire, profile, err := h.repository.Submit(c.Request.Context(), userID, input.Mode, input.Answers, generated)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"questionnaire": questionnaire, "profile": profile}})
}

func (h *Handler) CurrentProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	profile, err := h.repository.GetProfile(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"profile": profile}})
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(auth.ContextUserIDKey))
	return id, err == nil
}

func writeRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUnavailable):
		writeError(c, http.StatusServiceUnavailable, "service_unavailable", "questionnaire service unavailable")
	case errors.Is(err, ErrProfileNotFound):
		writeError(c, http.StatusNotFound, "profile_not_found", "profile not found")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
