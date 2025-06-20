package logx

import (
	"context"
	"fmt"
	"github.com/Privasea/tl/injection"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"strings"
	"time"
)

var dl = true
var dlw = true
var dlr = true

const (
	normalSql = "normal"
	slowSql   = "slow"
	errSql    = "error"
	warnSql   = "warn"
)

func Default(level logger.LogLevel) logger.Interface {
	var cfg = logger.Config{
		SlowThreshold: 100 * time.Millisecond,
		LogLevel:      level,
		Colorful:      true,
	}

	return &gormLogger{
		Config: cfg,
	}
}

type gormLogger struct {
	logger.Config
}

// LogMode logger mode
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	begin, end := time.Now(), time.Now()
	if l.LogLevel >= logger.Info {
		errInfo := fmt.Sprintf(msg, data)
		gormWriter(ctx, LevelWarn, 0, "", "", utils.FileWithLineNum(), normalSql, errInfo, begin, end)
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	begin, end := time.Now(), time.Now()
	if l.LogLevel >= logger.Warn {
		errInfo := fmt.Sprintf(msg, data)
		gormWriter(ctx, LevelWarn, 0, "", "", utils.FileWithLineNum(), warnSql, errInfo, begin, end)
	}
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	begin, end := time.Now(), time.Now()
	if l.LogLevel >= logger.Error {
		errInfo := fmt.Sprintf(msg, data)
		gormWriter(ctx, LevelError, 0, "", "", utils.FileWithLineNum(), errSql, errInfo, begin, end)
	}
}

// Trace print sql message
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	if l.LogLevel > 0 {
		end := time.Now()
		elapsed := end.Sub(begin)
		switch {
		case err != nil && l.LogLevel >= logger.Error:
			sql, rows := fc()
			if rows == -1 {
				gormWriter(ctx, LevelError, rows, sql, "", utils.FileWithLineNum(), slowSql, errMsg, begin, end)
			} else {
				gormWriter(ctx, LevelError, rows, sql, "", utils.FileWithLineNum(), normalSql, errMsg, begin, end)
			}
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				gormWriter(ctx, LevelWarn, rows, sql, slowLog, utils.FileWithLineNum(), slowSql, errMsg, begin, end)
			} else {
				gormWriter(ctx, LevelWarn, rows, sql, "", utils.FileWithLineNum(), normalSql, errMsg, begin, end)
			}
		case l.LogLevel >= logger.Info:
			sql, rows := fc()
			gormWriter(ctx, LevelInfo, rows, sql, "", utils.FileWithLineNum(), normalSql, errMsg, begin, end)
		}
	}
}

func gormWriter(ctx context.Context, level string, rows int64, sql, slowLog, line, tag, errMsg string, begin, end time.Time) {

	var requestId string

	if ginCtx, ok := ctx.(*gin.Context); ok {
		requestId = ginCtx.GetString(xB3Key)
		if requestId == "" {
			requestId = ginCtx.GetHeader(xB3Key)
		}
	}
	if requestId == "" {
		requestId = "null"
	}
	
	database, ok := ctx.Value("gorm:database").(string)
	if !ok {
		database = "null"
	}
	msg := "db_log"
	request := gormRequestLog{
		Database: database,
		Rows:     rows,
		Sql:      sql,
		Tag:      tag,
		SlowLog:  slowLog,
		Line:     line,
		ErrMsg:   errMsg,
	}

	elapsed := end.Sub(begin)
	runTime := fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)

	var fields []zap.Field
	fields = append(fields,
		zap.Any("datetime", begin.Format(timeFormat)),
		zap.String("message_type", "dblog"),
		zap.String(requestIdKey, requestId),
		zap.String(trackingPoints, injection.GetTrackingPoint(ctx.(*gin.Context))),
		zap.Any("request", request),
		zap.String("respon", errMsg),
		zap.Any("start_time", float64(begin.UnixNano())/1e9),
		zap.Any("end_time", float64(end.UnixNano())/1e9),
		zap.String("runtime", runTime),
	)
	//关闭所有sql日志
	if !dl {
		return
	}
	//区分读和写操作的
	if containsSelect(sql) {
		//关闭读操作日志
		if !dlr {
			return
		}
	} else {
		//关闭写操作日志
		if !dlw {
			return
		}
	}

	switch level {
	case LevelInfo:

		getSugar().Info(msg, fields...)

	case LevelWarn:

		getSugar().Warn(msg, fields...)

	case LevelError:

		getSugar().Error(msg, fields...)

	}

	return
}

type gormRequestLog struct {
	Database string `json:"database"`
	Rows     int64  `json:"rows"`
	Sql      string `json:"sql"`
	Tag      string `json:"tag"`
	SlowLog  string `json:"slow_log"`
	Line     string `json:"line"`
	ErrMsg   string `json:"err_msg"`
}

func SetDbLog(v bool) {
	dl = v
}
func SetDbWLog(v bool) {
	dlw = v
}
func SetDbRLog(v bool) {
	dlr = v
}
func containsSelect(sql string) bool {
	sql = strings.ToLower(sql) // 转换为小写
	return strings.Contains(sql, "select")
}
