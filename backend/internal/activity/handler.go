package activity

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

type createActivityRequest struct {
	Title         string     `json:"title"`
	Type          string     `json:"type"`
	Description   string     `json:"description"`
	RequiredCount int        `json:"required_count"`
	Tags          StringList `json:"tags"`
	PreferredTags StringList `json:"preferred_tags"`
	TimeText      string     `json:"time_text"`
	LocationText  string     `json:"location_text"`
}

type applyRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	var input createActivityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Type = strings.TrimSpace(input.Type)
	input.TimeText = strings.TrimSpace(input.TimeText)
	input.LocationText = strings.TrimSpace(input.LocationText)
	if input.Title == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "title is required")
		return
	}
	if input.Description == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "description is required")
		return
	}
	if input.RequiredCount <= 0 {
		input.RequiredCount = 2
	}
	if input.Type == "" {
		input.Type = "project"
	}
	model := Activity{
		CreatorID:     userID,
		Title:         input.Title,
		Type:          input.Type,
		Description:   input.Description,
		RequiredCount: input.RequiredCount,
		JoinedCount:   0,
		Tags:          normalizeList(input.Tags),
		PreferredTags: normalizeList(input.PreferredTags),
		TimeText:      input.TimeText,
		LocationText:  input.LocationText,
		Status:        StatusRecruiting,
	}
	if err := h.repository.CreateActivity(c.Request.Context(), &model); err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"activity": model}})
}

func (h *Handler) List(c *gin.Context) {
	activities, err := h.repository.ListActivities(c.Request.Context(), ListFilter{
		Type:    strings.TrimSpace(c.Query("type")),
		Status:  strings.TrimSpace(c.Query("status")),
		Keyword: strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"activities": activities}})
}

func (h *Handler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	activity, err := h.repository.GetActivity(c.Request.Context(), id)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"activity": activity}})
}

func (h *Handler) MyActivities(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	activities, err := h.repository.ListMyActivities(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"activities": activities}})
}

func (h *Handler) Apply(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	activityID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var input applyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	model := Application{
		ActivityID:  activityID,
		ApplicantID: userID,
		Reason:      strings.TrimSpace(input.Reason),
		MatchScore:  0,
		Status:      ApplicationPending,
	}
	if err := h.repository.CreateApplication(c.Request.Context(), &model); err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"application": model}})
}

func (h *Handler) MyApplications(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	applications, err := h.repository.ListMyApplications(c.Request.Context(), userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"applications": applications}})
}

func (h *Handler) ActivityApplications(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	activityID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	applications, err := h.repository.ListActivityApplications(c.Request.Context(), activityID, userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"applications": applications}})
}

func (h *Handler) Approve(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	applicationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	application, err := h.repository.ApproveApplication(c.Request.Context(), applicationID, userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"application": application}})
}

func (h *Handler) Reject(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	applicationID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	application, err := h.repository.RejectApplication(c.Request.Context(), applicationID, userID)
	if err != nil {
		writeRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"application": application}})
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(auth.ContextUserIDKey))
	return id, err == nil
}

func parseIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_id", "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

func normalizeList(values []string) StringList {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return StringList(result)
}

func writeRepositoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUnavailable):
		writeError(c, http.StatusServiceUnavailable, "service_unavailable", "activity service unavailable")
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrApplicationNotFound):
		writeError(c, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, ErrDuplicateApply):
		writeError(c, http.StatusConflict, "duplicate_application", "cannot apply to the same activity twice")
	case errors.Is(err, ErrForbidden):
		writeError(c, http.StatusForbidden, "forbidden", "permission denied")
	case errors.Is(err, ErrInvalidState):
		writeError(c, http.StatusBadRequest, "invalid_state", "operation is not allowed in current state")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
