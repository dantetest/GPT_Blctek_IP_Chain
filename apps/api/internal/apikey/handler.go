package apikey

import (
	"errors"
	"net/http"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/principal"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	Name      string     `json:"name" binding:"required,max=120"`
	Scopes    []string   `json:"scopes" binding:"required,min=1"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (h *Handler) Create(c *gin.Context) {
	current, ok := principal.Get(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	var request createRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid api key request")
		return
	}
	created, err := h.service.Create(c.Request.Context(), CreateInput{
		UserID:    current.UserID,
		Name:      request.Name,
		Scopes:    request.Scopes,
		ExpiresAt: request.ExpiresAt,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "API_KEY_CREATED", "api key created; store the secret now", created)
}

func (h *Handler) List(c *gin.Context) {
	current, ok := principal.Get(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	keys, err := h.service.List(c.Request.Context(), current.UserID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "API_KEY_LIST_FAILED", "could not list api keys")
		return
	}
	httpx.OK(c, "API_KEYS", keys)
}

func (h *Handler) Context(c *gin.Context) {
	current, ok := principal.Get(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "api key authentication is required")
		return
	}
	httpx.OK(c, "API_KEY_CONTEXT", gin.H{
		"user_id":    current.UserID,
		"api_key_id": current.APIKeyID,
		"tier":       current.Tier,
		"scopes":     current.Scopes,
	})
}

func (h *Handler) Revoke(c *gin.Context) {
	current, ok := principal.Get(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	if err := h.service.Revoke(c.Request.Context(), current.UserID, c.Param("id")); err != nil {
		h.handleError(c, err)
		return
	}
	httpx.OK(c, "API_KEY_REVOKED", gin.H{})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotAvailable):
		httpx.Error(c, http.StatusForbidden, "API_KEYS_NOT_AVAILABLE", "upgrade to professional or enterprise to create api keys")
	case errors.Is(err, ErrLimitReached):
		httpx.Error(c, http.StatusConflict, "API_KEY_LIMIT_REACHED", "active api key limit reached")
	case errors.Is(err, ErrInvalidName):
		httpx.Error(c, http.StatusUnprocessableEntity, "INVALID_API_KEY_NAME", "api key name is invalid")
	case errors.Is(err, ErrInvalidScope):
		httpx.Error(c, http.StatusUnprocessableEntity, "INVALID_API_KEY_SCOPE", "one or more scopes are invalid")
	case errors.Is(err, ErrKeyNotFound):
		httpx.Error(c, http.StatusNotFound, "API_KEY_NOT_FOUND", "api key was not found")
	case errors.Is(err, ErrAccountNotActive):
		httpx.Error(c, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE", "account is not active")
	default:
		httpx.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
