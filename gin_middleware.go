package tl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Privasea/tl/gcalx"
	"github.com/Privasea/tl/injection"
	"github.com/Privasea/tl/logx"
	"github.com/Privasea/tl/utils"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var paramLog = true
var NoParamLogUrlPath []string

func setParamLog(v bool) {
	paramLog = v
}
func SetNoParamLogByUrlPath(urlPath []string) {
	NoParamLogUrlPath = urlPath
}

// 判断请求是否为 multipart/form-data 类型
func isMultipartFormData(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "multipart/form-data")
}

// GinInterceptor 记录框架出入参, 开启链路追踪
func GinInterceptor(ctx *gin.Context) {

	startTime := time.Now()
	//记录是否经过此中间件，用于判定后面需要从请求头取数据
	ctx.Set("is_cross_middleware", 1)

	logId := ctx.GetHeader(xB3Key)
	if logId == "" {
		logId = Md5(uuid.NewV4().String())
	}
	ctx.Set(xB3Key, logId)
	ctx.Set(requestIdKey, logId)

	for _, k := range gcalx.B3headers {
		v := ctx.GetHeader(k)
		if v != "" {
			ctx.Set(k, v)
		}
	}
	xhuayuTraffictTag := ctx.GetHeader("x-huayu-traffic-tag")
	if xhuayuTraffictTag != "" {
		ctx.Set("x-huayu-traffic-tag", xhuayuTraffictTag)
	}

	traceId := ctx.GetHeader("trace_id")
	if traceId != "" {
		ctx.Set("trace_id", traceId)
	}

	var rl interface{}
	var rpl interface{}

	w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: ctx.Writer}
	ctx.Writer = w
	//不为文件上传，才记录请求的body
	if !isMultipartFormData(ctx.Request) {
		requestData, _ := ctx.GetRawData()
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestData))
		var in interface{}
		err := json.NewDecoder(bytes.NewBuffer(requestData)).Decode(&in)
		if err != nil {
			rl = string(requestData)
		} else {
			rl = in
		}
	} else {
		rl = ""
	}
	//start
	injection.TrackingPoints(ctx)
	imei := ctx.GetHeader("Imei")
	var point = injection.GetTrackingPoint(ctx)
	mock, err := injection.DealPoints(logx.GetConfig().AppName, ctx.Request.URL.Path, ctx.Request.Method, imei, point)
	if err != nil {

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code":      200500,
			"msg":       "exceptionErr" + err.Error(),
			"error":     "",
			"timestamp": utils.Millisec(time.Now()),
		})
		ctx.Abort()
		return
	}
	if mock != "" {
		var result map[string]interface{}
		err := json.Unmarshal([]byte(mock), &result)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}

		ctx.JSON(http.StatusOK, result)

		ctx.Abort()
		return
	}
	//end
	var request = req{
		URL:            ctx.Request.URL.String(),
		Method:         ctx.Request.Method,
		IP:             []string{ctx.ClientIP()},
		Path:           ctx.Request.URL.Path,
		TrackingPoints: point,
		Headers:        ctx.Request.Header,
		Query:          ctx.Request.URL.RawQuery,
		Body:           rl,
	}

	ctx.Next()

	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	runTime := fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
	// 检查是否需要告警
	shouldAlertTimeout := false
	shouldAlertError := false
	alertType := ""
	alertReason := ""

	// 1. 检查超时
	if alertConfig.TimeoutEnabled && elapsed > alertConfig.Threshold {
		shouldAlertTimeout = true
		alertType = "接口超时告警"
		alertReason = fmt.Sprintf("执行时间: %s, 超过阈值: %v", runTime, alertConfig.Threshold)
	}
	// 2. 检查HTTP状态码和业务状态码
	if alertConfig.ErrorEnabled && !containsString(alertConfig.IgnorePaths, ctx.Request.URL.Path) {
		statusCode := ctx.Writer.Status()
		if statusCode != http.StatusOK {
			shouldAlertError = true
			alertType = "接口状态码异常"
			alertReason = fmt.Sprintf("HTTP状态码: %d", statusCode)
		} else {
			// 尝试解析响应体
			var resp Response
			if err := json.Unmarshal(w.body.Bytes(), &resp); err == nil {
				if resp.Code != 0 {
					shouldAlertError = true
					alertType = "接口业务异常"
					alertReason = fmt.Sprintf("业务码: %d, 错误信息: %s", resp.Code, resp.Msg)
				}
			}
		}
	}
	// 发送告警-异常情况只发生一次
	if shouldAlertTimeout || shouldAlertError {
		// 获取 Redis 客户端并检查是否需要发送告警
		if redis := WrapRedis(ctx, alertConfig.RedisName); redis != nil {
			key := fmt.Sprintf("alarm:%s:%s", alertType, ctx.Request.URL.Path)
			// 尝试设置告警标记，5分钟内不重复
			if ok, _ := redis.SetNX(ctx, key, 1, 5*time.Minute).Result(); ok {
				go func() {
					msg := fmt.Sprintf("%s\n"+
						"- 告警原因：%s\n"+
						"- 请求路径：%s\n"+
						"- 请求方法：%s\n"+
						"- 执行时间：%s\n"+
						"- 客户端IP：%s\n"+
						"- 请求参数：%v\n"+
						"- 响应结果：%s",
						alertType,
						alertReason,
						ctx.Request.URL.Path,
						ctx.Request.Method,
						runTime,
						ctx.ClientIP(),
						request.Body,
						w.body.String())

					SendToFeishu(alertConfig.FeishuURL, alertConfig.AppName,alertConfig.AppEnv,alertType, msg, "no")
				}()
			}
		}
	}

	logW := true
	path := ctx.Request.URL.Path
	strSlice := []string{"/"}

	if utils.InArray(strSlice, path) {
		logW = false
	}
	bool := containsString(NoParamLogUrlPath, path) // 不在不需要记录的路由中的接口，才记录
	if paramLog && logW && bool == false {
		endTime := time.Now()
		elapsed := endTime.Sub(startTime)
		runTime := fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
		var out interface{}
		err := json.NewDecoder(w.body).Decode(&out)
		if err != nil {
			rpl = w.body.String()
		} else {
			rpl = out
		}

		logx.InfoF(ctx, "%s", "api_log",
			zap.Any("datetime", startTime.Format(timeFormat)),
			zap.Any("message_type", "apilog"),
			zap.Any("request", request),
			zap.Any("respon", rpl),
			zap.Any("start_time", float64(startTime.UnixNano())/1e9),
			zap.Any("end_time", float64(endTime.UnixNano())/1e9),
			zap.String("runtime", runTime),
		)
	}
}
func containsString(slice []string, target string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[target]
	return ok
}

// GinInterceptor 记录框架出入参, 开启链路追踪
func GinInterceptorJob(ctx *gin.Context) {

	ctx.Set("is_cross_middleware", 1) //记录是否经过此中间件，用于判定后面需要从请求头取数据
	logId := Md5(uuid.NewV4().String())
	ctx.Set(xB3Key, logId)
	ctx.Set(requestIdKey, logId)

}
func GinInterceptorJobByTask(ctx *gin.Context, taskName string) {

	ctx.Set("is_cross_middleware", 1) //记录是否经过此中间件，用于判定后面需要从请求头取数据
	logId := taskName + "-" + Md5(uuid.NewV4().String())
	ctx.Set(xB3Key, logId)
	ctx.Set(requestIdKey, logId)

}

type req struct {
	URL            string              `json:"url"`
	Method         string              `json:"method"`
	IP             []string            `json:"ip"`
	Path           string              `json:"path"`
	Headers        map[string][]string `json:"headers"`
	Query          interface{}         `json:"query"`
	Body           interface{}         `json:"body"`
	TrackingPoints string              `json:"tracking_points"`
}
