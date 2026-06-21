package admin

import (
	"errors"
	"net/http"
	"strconv"

	"matchlab/backend/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(repository Repository) *Handler { return NewHandlerWithService(NewService(repository)) }

func NewHandlerWithService(service Service) *Handler { return &Handler{service: service} }

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.service.Stats(c.Request.Context())
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"stats": stats}})
}

func (h *Handler) Users(c *gin.Context) {
	page, ok := pageFromQuery(c)
	if !ok {
		return
	}
	users, err := h.service.Users(c.Request.Context(), UsersFilter{
		Keyword: c.Query("keyword"), Role: c.Query("role"), Page: page,
	})
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"users": users}})
}

func (h *Handler) Activities(c *gin.Context) {
	page, ok := pageFromQuery(c)
	if !ok {
		return
	}
	activities, err := h.service.Activities(c.Request.Context(), ActivitiesFilter{
		Keyword: c.Query("keyword"), Type: c.Query("type"), Status: c.Query("status"), Page: page,
	})
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"activities": activities}})
}

func (h *Handler) Applications(c *gin.Context) {
	page, ok := pageFromQuery(c)
	if !ok {
		return
	}
	applications, err := h.service.Applications(c.Request.Context(), ApplicationsFilter{
		Status: c.Query("status"), Page: page,
	})
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"applications": applications}})
}

func (h *Handler) Feedbacks(c *gin.Context) {
	page, ok := pageFromQuery(c)
	if !ok {
		return
	}
	feedbacks, err := h.service.Feedbacks(c.Request.Context(), page)
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"feedbacks": feedbacks}})
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	requesterID, err := uuid.Parse(c.GetString(auth.ContextUserIDKey))
	if err != nil {
		writeJSONError(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
		return
	}
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid_request", "invalid user id")
		return
	}
	var input struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}
	updated, err := h.service.UpdateUserRole(c.Request.Context(), requesterID, targetID, input.Role)
	if err != nil {
		writeHandlerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": updated}})
}

func pageFromQuery(c *gin.Context) (Page, bool) {
	page := Page{}
	if raw, exists := c.GetQuery("limit"); exists {
		value, err := strconv.Atoi(raw)
		if err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request", "limit must be an integer")
			return Page{}, false
		}
		page.Limit = value
	}
	if raw, exists := c.GetQuery("offset"); exists {
		value, err := strconv.Atoi(raw)
		if err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request", "offset must be an integer")
			return Page{}, false
		}
		page.Offset = value
	}
	return page, true
}

func writeHandlerError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidFilter), errors.Is(err, ErrInvalidRole):
		writeJSONError(c, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, ErrSelfDemotion):
		writeJSONError(c, http.StatusBadRequest, "self_demotion_forbidden", err.Error())
	case errors.Is(err, ErrNotFound):
		writeJSONError(c, http.StatusNotFound, "not_found", "user not found")
	case errors.Is(err, ErrUnavailable):
		writeJSONError(c, http.StatusServiceUnavailable, "service_unavailable", "admin service unavailable")
	default:
		writeJSONError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeJSONError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
