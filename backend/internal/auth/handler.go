package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"matchlab/backend/internal/user"
)

const (
	// ContextUserIDKey is populated by JWT middleware.
	ContextUserIDKey = "user_id"
	// ContextRoleKey is populated by JWT middleware.
	ContextRoleKey = "role"
)

// Handler exposes authentication HTTP endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates authentication HTTP handlers.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register creates a user account.
func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	registered, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"user": registered}})
}

// Login verifies credentials and returns an access token.
func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求格式不正确")
		return
	}
	result, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// Me returns the currently authenticated user.
func (h *Handler) Me(c *gin.Context) {
	id, err := uuid.Parse(c.GetString(ContextUserIDKey))
	if err != nil {
		writeError(c, http.StatusUnauthorized, "invalid_token", "登录凭证无效或已过期")
		return
	}
	current, err := h.service.CurrentUser(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": current}})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(c, http.StatusBadRequest, "invalid_request", "邮箱格式不正确或密码长度不符合要求")
	case errors.Is(err, ErrInvalidCredentials):
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误")
	case errors.Is(err, user.ErrEmailExists):
		writeError(c, http.StatusConflict, "email_exists", "该邮箱已注册")
	case errors.Is(err, user.ErrNotFound):
		writeError(c, http.StatusNotFound, "user_not_found", "用户不存在")
	case errors.Is(err, user.ErrUnavailable):
		writeError(c, http.StatusServiceUnavailable, "service_unavailable", "认证服务暂不可用")
	default:
		writeError(c, http.StatusInternalServerError, "internal_error", "服务器内部错误")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
