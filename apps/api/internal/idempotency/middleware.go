package idempotency

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/httpx"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/principal"
	"github.com/gin-gonic/gin"
)

const (
	headerName     = "Idempotency-Key"
	maxRequestBody = 2 << 20
)

type captureWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (writer *captureWriter) Write(data []byte) (int, error) {
	writer.body.Write(data)
	return writer.ResponseWriter.Write(data)
}

func (writer *captureWriter) WriteString(value string) (int, error) {
	writer.body.WriteString(value)
	return writer.ResponseWriter.WriteString(value)
}

func Require(store *Store, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader(headerName))
		if len(key) < 8 || len(key) > 160 {
			httpx.Error(c, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", "Idempotency-Key must contain 8-160 characters")
			return
		}
		current, ok := principal.Get(c)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required")
			return
		}

		body, err := readRequestBody(c)
		if err != nil {
			httpx.Error(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "request body exceeds idempotency limit")
			return
		}
		requestHash := hashRequest(c.Request.Method, c.FullPath(), current.UserID, body)
		scope := c.Request.Method + ":" + c.FullPath() + ":" + current.UserID
		record, created, err := store.Begin(c.Request.Context(), scope, key, requestHash, time.Now().UTC().Add(ttl))
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, "IDEMPOTENCY_UNAVAILABLE", "idempotency service is unavailable")
			return
		}
		if !created {
			if record.RequestHash != requestHash {
				httpx.Error(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency-Key was already used for a different request")
				return
			}
			if record.ResponseStatus == nil {
				httpx.Error(c, http.StatusConflict, "IDEMPOTENCY_IN_PROGRESS", "request with this Idempotency-Key is still in progress")
				return
			}
			c.Header("Idempotency-Replayed", "true")
			c.Data(*record.ResponseStatus, "application/json; charset=utf-8", record.ResponseBody)
			c.Abort()
			return
		}

		writer := &captureWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()

		status := c.Writer.Status()
		responseBody := writer.body.Bytes()
		if status >= http.StatusInternalServerError || !json.Valid(responseBody) {
			_ = store.Delete(c.Request.Context(), record.ID)
			return
		}
		if err := store.Complete(c.Request.Context(), record.ID, status, responseBody); err != nil {
			_ = store.Delete(c.Request.Context(), record.ID)
		}
	}
}

func readRequestBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	limited := io.LimitReader(c.Request.Body, maxRequestBody+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(body) > maxRequestBody {
		return nil, ErrRequestTooLarge
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func hashRequest(method, path, userID string, body []byte) string {
	hash := sha256.New()
	hash.Write([]byte(method))
	hash.Write([]byte{0})
	hash.Write([]byte(path))
	hash.Write([]byte{0})
	hash.Write([]byte(userID))
	hash.Write([]byte{0})
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}
