package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
)
const requestIDHeader="X-Request-ID"
func RequestID() gin.HandlerFunc { return func(c *gin.Context){ requestID:=c.GetHeader(requestIDHeader); if requestID==""{ requestID=newRequestID() }; c.Set("request_id",requestID); c.Header(requestIDHeader,requestID); c.Next() } }
func newRequestID() string { bytes:=make([]byte,12); if _,err:=rand.Read(bytes); err!=nil{return "req_unavailable"}; return "req_"+hex.EncodeToString(bytes) }
