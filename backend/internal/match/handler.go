package match

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/activity"
	"matchlab/backend/internal/auth"
)

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{repository: repository}
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
	if strings.ToLower(strings.TrimSpace(input.TargetType)) != "activity" {
		writeError(c, http.StatusBadRequest, "invalid_request", "target_type must be activity")
		return
	}
	limit := 10
	if input.Limit != nil {
		limit = *input.Limit
	}
	if limit < 1 || limit > 50 {
		writeError(c, http.StatusBadRequest, "invalid_request", "limit must be between 1 and 50")
		return
	}
	signals, err := h.repository.LoadSignals(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	candidates, err := h.repository.ListCandidates(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	eligible := make([]activity.Activity, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.CreatorID == userID || (candidate.Status != "" && candidate.Status != activity.StatusRecruiting) {
			continue
		}
		eligible = append(eligible, candidate)
	}
	recommendations := RankActivities(signals, eligible, limit)
	if err := h.repository.UpsertMatches(c.Request.Context(), userID, signals.QuestionnaireID, recommendations); err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"recommendations": recommendations}})
}

func (h *Handler) CurrentMatches(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	matches, err := h.repository.ListMatches(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"matches": matches}})
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(auth.ContextUserIDKey))
	return id, err == nil
}

func writeRepositoryError(c *gin.Context, err error) {
	switch {
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
