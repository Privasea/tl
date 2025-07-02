package tl

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"time"
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

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Enabled   bool          // 是否启用超时告警
	Threshold time.Duration // 超时阈值（例如：6 * time.Second）
	FeishuURL string        // 飞书webhook URL
	AppName string
	AppEnv string
	RedisName string        // Redis实例名称，默认为"default"
}

var timeoutConfig = TimeoutConfig{
	Enabled:   false,
	Threshold: 6 * time.Second, // 默认6秒
	RedisName: "default",
	FeishuURL: "",
	AppName: "test",
	AppEnv: "dev",
}

// SetTimeoutConfig 设置超时配置
func SetTimeoutConfig(config TimeoutConfig) {
	timeoutConfig = config
}

// GetTimeoutConfig 获取当前超时配置
func GetTimeoutConfig() TimeoutConfig {
	return timeoutConfig
}

// EnableTimeout 启用超时告警
func EnableTimeout(threshold time.Duration, feishuURL string) {
	timeoutConfig.Enabled = true
	timeoutConfig.Threshold = threshold
	timeoutConfig.FeishuURL = feishuURL
}

// DisableTimeout 禁用超时告警
func DisableTimeout() {
	timeoutConfig.Enabled = false
}
