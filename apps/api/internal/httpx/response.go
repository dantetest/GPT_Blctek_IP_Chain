package httpx

import "github.com/gin-gonic/gin"
type Envelope struct { Code string `json:"code"`; Message string `json:"message"`; Data any `json:"data,omitempty"`; RequestID string `json:"request_id"` }
func OK(c *gin.Context, code string, data any) { requestID,_:=c.Get("request_id"); c.JSON(200,Envelope{Code:code,Message:"success",Data:data,RequestID:stringValue(requestID)}) }
func Error(c *gin.Context,status int,code,message string){ requestID,_:=c.Get("request_id"); c.AbortWithStatusJSON(status,Envelope{Code:code,Message:message,RequestID:stringValue(requestID)}) }
func stringValue(value any) string { if result,ok:=value.(string); ok{return result}; return "" }
