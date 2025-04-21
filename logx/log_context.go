package logx

import "github.com/gin-gonic/gin"

func InfoF(ctx *gin.Context, template string, args ...interface{}) {
	logId := ctx.GetString(xB3Key)
	if logId == "" && ctx.GetString("is_cross_middleware") == "" {
		logId = ctx.GetHeader(xB3Key)
	}
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelInfo, msg, fields...)
}

func WarnF(ctx *gin.Context, template string, args ...interface{}) {
	logId := ctx.GetString(xB3Key)
	if logId == "" && ctx.GetString("is_cross_middleware") == "" {
		logId = ctx.GetHeader(xB3Key)
	}
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelWarn, msg, fields...)
}

// ErrorF 打印程序错误日志
func ErrorF(ctx *gin.Context, template string, args ...interface{}) {
	logId := ctx.GetString(xB3Key)
	if logId == "" && ctx.GetString("is_cross_middleware") == "" {
		logId = ctx.GetHeader(xB3Key)
	}
	msg, fields := dealWithArgs(template, args...)
	writer(logId, LevelError, msg, fields...)
}
