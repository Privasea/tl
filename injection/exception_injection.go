package injection

import (
	"context"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

var dl = false

func SetLogErr(v bool) {
	dl = v
}
func TrackingPoints(ctx *gin.Context) {
	trackingPoints := ctx.GetString("tracking_points")
	if trackingPoints == "" {
		ctx.Set("tracking_points", "1001")
	} else {
		num, _ := strconv.Atoi(trackingPoints)
		ctx.Set("tracking_points", strconv.Itoa(num+1))
	}

}
func GetTrackingPoint(ctx context.Context) string {
	var trackingPoint string
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if ginCtx.Request == nil {
			return ""
		}
		trackingPoint = ginCtx.GetString("tracking_points")
		if trackingPoint == "" {
			trackingPoint = ginCtx.GetHeader("tracking_points")
		}
	}

	return trackingPoint
}
func GetMatchPoints(server, path, method, imei, trackingPoints string) (LogItem, error) {

	datas, err := GetApiDataForInject(server, path, method, imei)
	if err == nil {
		for _, data := range datas {
			if strconv.Itoa(data.PointCode) == trackingPoints && imei != "" {
				return data, nil
			}
		}
	}

	return LogItem{}, nil

}

func DealPoints(server, path, method, imei, trackingPoints string) (string, error) {

	pointData, err := GetMatchPoints(server, path, method, imei, trackingPoints)
	if err != nil {
		return "", err
	}
	if pointData.ID != 0 {
		if pointData.InjectionType == "sleep" {
			// 将字符串转换为整数
			seconds, err := strconv.Atoi(pointData.InjectionData)
			if err != nil {
				return "", err
			}
			time.Sleep(time.Duration(seconds) * time.Second)
		}
		if pointData.InjectionType == "panic" {
			panic(pointData.InjectionData)
		}
		if pointData.InjectionType == "mock" {
			return pointData.InjectionData, nil
		}
		if pointData.InjectionType == "err" {
			return pointData.InjectionData, nil
		}
	}
	return "", nil
}
