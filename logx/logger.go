// Package logx
// logx: this is extend package, use https://github.com/uber-go/zap
package logx

import (
	"facebyte/pkg/tl/logx/logger"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"time"
)

const (
	envLogPath       = "/Users/zhangjiulin/facebytelog/"
	defaultChildPath = "file-%Y-%m-%d.log" // 默认子目录

	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

var (
	appName string
	appMode string
	logType string
	sugar   *zap.Logger
)

func SetLogger(name, mode, t string) {
	appName = name
	appMode = mode
	logType = t
}

type config struct {
	appName   string // 应用名
	appMode   string // 应用环境
	logType   string // 日志类型
	logPath   string // 日志主路径
	childPath string // 日志子路径+文件名
}

func getSugar() *zap.Logger {
	if sugar == nil {
		cfg := config{
			appName:   appName,
			appMode:   appMode,
			logType:   logType,
			logPath:   envLogPath,
			childPath: defaultChildPath,
		}

		if appName == "" {
			cfg.appName = "default_app"
		}
		if appMode == "" {
			cfg.appMode = "dev1"
		}

		sugar = initSugar(&cfg)
	}

	return sugar
}

func initSugar(lc *config) *zap.Logger {
	loglevel := zapcore.InfoLevel
	defaultLogLevel := zap.NewAtomicLevel()
	defaultLogLevel.SetLevel(loglevel)

	logPath := fmt.Sprintf("%s/%s/%s", lc.logPath, lc.appName, lc.childPath)

	var core zapcore.Core
	// 打印至文件中
	if lc.logType == "file" {
		configs := zap.NewProductionEncoderConfig()
		configs.FunctionKey = "func"
		configs.EncodeTime = timeEncoder

		w := zapcore.AddSync(GetWriter(logPath))

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(configs),
			w,
			defaultLogLevel,
		)
		log.Printf("[app] logger success")
	} else {
		// 打印在控制台
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), defaultLogLevel)
		log.Printf("[app] logger success")
	}

	filed := zap.Fields(zap.String("app_name", lc.appName), zap.String("app_mode", lc.appMode))
	return zap.New(core, filed, zap.AddCaller(), zap.AddCallerSkip(3))
}

func dealWithArgs(tmp string, args ...interface{}) (msg string, f []zap.Field) {
	if len(args) > 0 {
		var tmpArgs []interface{}
		for _, item := range args {
			if nil == item {
				continue
			}
			if zapField, ok := item.(zap.Field); ok {
				f = append(f, zapField)
			} else {
				tmpArgs = append(tmpArgs, item)
			}
		}
		if len(tmpArgs) > 0 {
			msg = fmt.Sprintf(tmp, tmpArgs...)
		}
	} else {
		msg = tmp
	}
	return
}

func writer(logId, level, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(requestIdKey, logId))

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

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	var layout = "2006-01-02 15:04:05"
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// GetWriter 按天切割按大小切割
func GetWriter(filename string) io.Writer {
	hook, err := logger.New(filename)

	if err != nil {
		panic(err)
	}
	return hook
}
