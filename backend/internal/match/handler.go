package match

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

type recommendRequest struct {
	TargetType string `json:"target_type"`
	Limit      *int   `json:"limit"`
}

func (h *Handler) Recommend(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	var input recommendRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	recommendations, err := h.service.Recommend(c.Request.Context(), userID, input.TargetType, input.Limit)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"recommendations": recommendations}})
}

func (h *Handler) MyMatches(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	matches, err := h.service.MyMatches(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"matches": matches}})
}

// CurrentMatches is kept as a compatibility alias while callers migrate to MyMatches.
func (h *Handler) CurrentMatches(c *gin.Context) { h.MyMatches(c) }

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(auth.ContextUserIDKey))
	return id, err == nil
}

func writeRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidTarget), errors.Is(err, ErrInvalidLimit):
		writeError(c, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, ErrUnavailable):
		writeError(c, http.StatusServiceUnavailable, "service_unavailable", "match service unavailable")
	case errors.Is(err, ErrProfileRequired):
		writeError(c, http.StatusBadRequest, "profile_required", "submit a questionnaire before requesting recommendations")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
