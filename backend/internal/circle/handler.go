package circle

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/auth"
)

type Handler struct{ service Service }
func NewHandler(r Repository) *Handler { return &Handler{service: NewService(r)} }
func userID(c *gin.Context) (uuid.UUID, bool) { id, err := uuid.Parse(c.GetString(auth.ContextUserIDKey)); return id, err == nil }

func idParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil { respondError(c, http.StatusBadRequest, "invalid_id", "invalid id"); return uuid.Nil, false }
	return id, true
}
func (h *Handler) List(c *gin.Context) { rows, err := h.service.ListApproved(c.Request.Context()); if err != nil { handleError(c, err); return }; c.JSON(http.StatusOK, gin.H{"data": gin.H{"circles": rows}}) }
func (h *Handler) Detail(c *gin.Context) { id, ok := idParam(c, "id"); if !ok { return }; row, err := h.service.GetApproved(c.Request.Context(), id); if err != nil { handleError(c, err); return }; c.JSON(http.StatusOK, gin.H{"data": gin.H{"circle": row}}) }
func (h *Handler) Members(c *gin.Context) { id, ok := idParam(c, "id"); if !ok { return }; rows, err := h.service.Members(c.Request.Context(), id); if err != nil { handleError(c, err); return }; c.JSON(http.StatusOK, gin.H{"data": gin.H{"members": rows}}) }
func (h *Handler) Create(c *gin.Context) {
	uid, ok := userID(c); if !ok { return }
	var in struct{ Name string; Description string; Category string }
	if c.ShouldBindJSON(&in) != nil { respondError(c, 400, "invalid_request", "invalid request body"); return }
	circle, channel, err := h.service.Create(c.Request.Context(), uid, in.Name, in.Description, in.Category)
	if err != nil { handleError(c, err); return }
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"circle": circle, "channel": channel}})
}
func (h *Handler) Join(c *gin.Context) { uid, _ := userID(c); id, ok := idParam(c, "id"); if !ok { return }; if err := h.service.Join(c.Request.Context(), id, uid); err != nil { handleError(c, err); return }; c.JSON(http.StatusOK, gin.H{"data": gin.H{"joined": true}}) }
func (h *Handler) Mine(c *gin.Context) { uid, _ := userID(c); rows, err := h.service.Mine(c.Request.Context(), uid); if err != nil { handleError(c, err); return }; c.JSON(200, gin.H{"data": gin.H{"circles": rows}}) }
func (h *Handler) Channels(c *gin.Context) { uid, _ := userID(c); cid, ok := idParam(c, "id"); if !ok { return }; rows, err := h.service.Channels(c.Request.Context(), cid, uid); if err != nil { handleError(c, err); return }; c.JSON(200, gin.H{"data": gin.H{"channels": rows}}) }
func (h *Handler) Messages(c *gin.Context) { uid, _ := userID(c); cid, ok := idParam(c, "id"); if !ok { return }; chid, ok := idParam(c, "channelId"); if !ok { return }; rows, err := h.service.Messages(c.Request.Context(), cid, chid, uid); if err != nil { handleError(c, err); return }; c.JSON(200, gin.H{"data": gin.H{"messages": rows}}) }
func (h *Handler) PostMessage(c *gin.Context) {
	uid, _ := userID(c); cid, ok := idParam(c, "id"); if !ok { return }; chid, ok := idParam(c, "channelId"); if !ok { return }
	var in struct{ Content string }
	if c.ShouldBindJSON(&in) != nil { respondError(c, 400, "invalid_request", "invalid request body"); return }
	message, err := h.service.PostMessage(c.Request.Context(), cid, chid, uid, in.Content)
	if err != nil { handleError(c, err); return }
	c.JSON(201, gin.H{"data": gin.H{"message": message}})
}
func (h *Handler) AdminList(c *gin.Context) { status := strings.TrimSpace(c.Query("status")); rows, err := h.service.AdminList(c.Request.Context(), AdminFilter{Status: status}); if err != nil { handleError(c, err); return }; c.JSON(200, gin.H{"data": gin.H{"circles": rows}}) }
func (h *Handler) Approve(c *gin.Context) { h.moderate(c, "approved") }
func (h *Handler) Reject(c *gin.Context) { h.moderate(c, "rejected") }
func (h *Handler) moderate(c *gin.Context, status string) { id, ok := idParam(c, "id"); if !ok { return }; row, err := h.service.Moderate(c.Request.Context(), id, status); if err != nil { handleError(c, err); return }; c.JSON(200, gin.H{"data": gin.H{"circle": row}}) }
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUnavailable): respondError(c, 503, "service_unavailable", "circle service unavailable")
	case errors.Is(err, ErrNotFound): respondError(c, 404, "not_found", "circle not found")
	case errors.Is(err, ErrForbidden): respondError(c, 403, "forbidden", "circle membership required")
	case errors.Is(err, ErrAlreadyMember): respondError(c, 409, "already_member", "already a circle member")
	case errors.Is(err, ErrInvalidState): respondError(c, 400, "invalid_state", "circle is not approved")
	case errors.Is(err, ErrInvalidName): respondError(c, 400, "invalid_request", err.Error())
	case errors.Is(err, ErrInvalidDescription): respondError(c, 400, "invalid_request", err.Error())
	case errors.Is(err, ErrInvalidCategory): respondError(c, 400, "invalid_request", err.Error())
	case errors.Is(err, ErrInvalidStatus): respondError(c, 400, "invalid_request", err.Error())
	default: respondError(c, 500, "internal_error", "internal server error")
	}
}
func respondError(c *gin.Context, status int, code, message string) { c.JSON(status, gin.H{"error": code, "message": message}) }

