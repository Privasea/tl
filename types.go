package tl

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

const (
	xB3Key       = "x-b3-traceid" // 日志key
	requestIdKey = "request_id"   // 日志key
	timeFormat   = "2006-01-02 15:04:05"
)

// Md5 md5
func Md5(s string) string {
	m := md5.Sum([]byte(s))
	return hex.EncodeToString(m[:])
}

func GetNewGinContext() *gin.Context {
	ctx := new(gin.Context)
	uid := uuid.NewV4().String()
	ctx.Request = &http.Request{
		Header: make(map[string][]string),
	}
	ctx.Request.Header.Set(xB3Key, uid)
	ctx.Request.Header.Set(requestIdKey, uid)
	ctx.Set(xB3Key, uid)
	ctx.Set(requestIdKey, uid)
	return ctx
}
