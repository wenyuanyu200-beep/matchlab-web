package questionnaire

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/auth"
)

type Handler struct {
	service Service
}

func NewHandler(repository Repository) *Handler {
	return NewHandlerWithService(NewService(repository))
}

func NewHandlerWithService(service Service) *Handler {
	return &Handler{service: service}
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
	questionnaire, profile, err := h.service.Submit(c.Request.Context(), userID, input.Mode, input.Answers)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"questionnaire": questionnaire, "profile": profile}})
}

func (h *Handler) Profile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	profile, err := h.service.Profile(c.Request.Context(), userID)
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
	case errors.Is(err, ErrInvalidMode):
		writeError(c, http.StatusBadRequest, "invalid_request", ErrInvalidMode.Error())
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
