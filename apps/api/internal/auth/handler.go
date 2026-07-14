package auth

import (
	"errors"
	"net/http"

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

type registerRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"max=120"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid registration request")
		return
	}
	user, err := h.service.Register(c.Request.Context(), RegisterInput{
		Email:       request.Email,
		Password:    request.Password,
		DisplayName: request.DisplayName,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "USER_REGISTERED", "registration completed", user)
}

func (h *Handler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid login request")
		return
	}
	pair, user, err := h.service.Login(c.Request.Context(), LoginInput{
		Email:     request.Email,
		Password:  request.Password,
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.OK(c, "AUTHENTICATED", gin.H{"tokens": pair, "user": user})
}

func (h *Handler) Refresh(c *gin.Context) {
	var request tokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required")
		return
	}
	pair, err := h.service.Refresh(c.Request.Context(), RefreshInput{
		RefreshToken: request.RefreshToken,
		UserAgent:    c.Request.UserAgent(),
		IPAddress:    c.ClientIP(),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.OK(c, "TOKEN_REFRESHED", pair)
}

func (h *Handler) Logout(c *gin.Context) {
	var request tokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required")
		return
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "LOGOUT_FAILED", "could not revoke session")
		return
	}
	httpx.OK(c, "LOGGED_OUT", gin.H{})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	var request verifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "verification token is required")
		return
	}
	if err := h.service.VerifyEmail(c.Request.Context(), request.Token); err != nil {
		h.handleError(c, err)
		return
	}
	httpx.OK(c, "EMAIL_VERIFIED", gin.H{})
}

func (h *Handler) Me(c *gin.Context) {
	current, ok := principal.Get(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
		return
	}
	user, err := h.service.GetUser(c.Request.Context(), current.UserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.OK(c, "CURRENT_USER", user)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		httpx.Error(c, http.StatusUnprocessableEntity, "INVALID_EMAIL", "email address is invalid")
	case errors.Is(err, ErrEmailExists):
		httpx.Error(c, http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "email is already registered")
	case errors.Is(err, ErrWeakPassword):
		httpx.Error(c, http.StatusUnprocessableEntity, "WEAK_PASSWORD", "password must be 12-128 characters and include upper, lower and numeric characters")
	case errors.Is(err, ErrInvalidCredential):
		httpx.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
	case errors.Is(err, ErrEmailNotVerified):
		httpx.Error(c, http.StatusForbidden, "EMAIL_NOT_VERIFIED", "email verification is required")
	case errors.Is(err, ErrAccountLocked):
		httpx.Error(c, http.StatusLocked, "ACCOUNT_LOCKED", "account is temporarily locked")
	case errors.Is(err, ErrAccountDisabled):
		httpx.Error(c, http.StatusForbidden, "ACCOUNT_DISABLED", "account is not active")
	case errors.Is(err, ErrExpiredToken):
		httpx.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "token has expired")
	case errors.Is(err, ErrInvalidToken):
		httpx.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "token is invalid")
	default:
		httpx.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
