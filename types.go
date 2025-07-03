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

// Response 通用响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// AlertConfig
type AlertConfig struct {
	//环境
	AppName string
	AppEnv  string
	// 超时告警配置
	TimeoutEnabled bool          // 是否启用超时告警
	Threshold      time.Duration // 超时阈值

	// 错误告警配置
	ErrorEnabled bool     // 是否启用错误告警
	IgnorePaths  []string // 忽略告警的路径列表

	// 通用配置
	FeishuURL string // 飞书webhook URL
	RedisName string // Redis实例名称
}

// 默认配置
var alertConfig = AlertConfig{
	TimeoutEnabled: false,
	Threshold:      6 * time.Second,
	ErrorEnabled:   false,
	IgnorePaths:    []string{},
	RedisName:      "default",
}

// SetAlertConfig
func SetAlertConfig(config AlertConfig) {
	alertConfig = config
}

// GetAlertConfig
func GetAlertConfig() AlertConfig {
	return alertConfig
}

func EnableAlertConfig(threshold time.Duration, feishuURL string) {
	alertConfig.TimeoutEnabled = true
	alertConfig.ErrorEnabled = true
	alertConfig.Threshold = threshold
	alertConfig.FeishuURL = feishuURL
}

func DisableAlter() {
	alertConfig.TimeoutEnabled = false
	alertConfig.ErrorEnabled = false
}
