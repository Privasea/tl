package injection

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var dl = false

func SetLogErr(v bool) {
	dl = v
}

// 定义压力类型
type PressureType string

const (
	PressureCPU    PressureType = "cpu"    // CPU压力
	PressureMemory PressureType = "memory" // 内存压力
	PressureDB     PressureType = "db"     // 数据库压力
)

// 系统级压力控制器
type SystemPressureController struct {
	mu sync.Mutex
	// 记录各种压力注入的状态
	pressures map[PressureType]struct {
		active   bool
		level    int // 压力等级 1-100
		duration time.Duration
	}
}

var (
	globalController = &SystemPressureController{
		pressures: make(map[PressureType]struct {
			active   bool
			level    int
			duration time.Duration
		}),
	}
)

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

// 添加压力注入类型
func (pc *SystemPressureController) InjectPressure(pType PressureType, level int, duration time.Duration) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.pressures[pType] = struct {
		active   bool
		level    int
		duration time.Duration
	}{
		active:   true,
		level:    level,
		duration: duration,
	}

	// 启动压力注入
	go pc.applyPressure(pType)
}

// 应用压力
func (pc *SystemPressureController) applyPressure(pType PressureType) {
	switch pType {
	case PressureCPU:
		pc.applyCPUPressure()
	case PressureMemory:
		pc.applyMemoryPressure()
	case PressureDB:
		pc.applyDBPressure()
	}
}

// CPU压力模拟
func (pc *SystemPressureController) applyCPUPressure() {
	pressure := pc.pressures[PressureCPU]
	numCPU := runtime.NumCPU()

	// 根据压力等级计算需要占用的CPU核心数
	numThreads := numCPU * pressure.level / 100

	for i := 0; i < numThreads; i++ {
		go func() {
			start := time.Now()
			for time.Since(start) < pressure.duration {
				// 执行密集计算
				for j := 0; j < 1000000; j++ {
					_ = j * j
				}
			}
		}()
	}
}

// 内存压力模拟
func (pc *SystemPressureController) applyMemoryPressure() {
	pressure := pc.pressures[PressureMemory]

	// 计算要分配的内存大小（以MB为单位）
	memorySize := pressure.level * 100 // 每个等级分配100MB

	var memoryBlocks [][]byte

	go func() {
		for i := 0; i < memorySize && pc.pressures[PressureMemory].active; i++ {
			// 分配1MB内存
			block := make([]byte, 1024*1024)
			memoryBlocks = append(memoryBlocks, block)
			time.Sleep(100 * time.Millisecond)
		}

		// 等待持续时间后释放内存
		time.Sleep(pressure.duration)
		memoryBlocks = nil
		runtime.GC()
	}()
}

// 数据库压力模拟
func (pc *SystemPressureController) applyDBPressure() {
	pressure := pc.pressures[PressureDB]

	// 模拟数据库连接池压力
	go func() {
		start := time.Now()
		for time.Since(start) < pressure.duration {
			// 模拟数据库连接延迟
			time.Sleep(time.Duration(pressure.level) * time.Millisecond)
		}
	}()
}

// 修改DealPoints函数以支持压力注入
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
		if pointData.InjectionType == "pressure" {
			// 解析压力参数，格式: "type:level:duration"
			// 例如: "cpu:80:30s" 表示CPU压力80%持续30秒
			params := strings.Split(pointData.InjectionData, ":")
			if len(params) != 3 {
				return "", errors.New("invalid pressure injection format")
			}

			pType := PressureType(params[0])
			level, _ := strconv.Atoi(params[1])          //压力等级:80%
			duration, _ := time.ParseDuration(params[2]) //持续时间

			// 判断是否是接口级注入
			if strings.HasPrefix(params[0], "api_") {
				// 接口级压力注入
				switch pType {
				case "api_cpu":
					// 只在当前请求中模拟CPU压力
					go func() {
						start := time.Now()
						for time.Since(start) < duration {
							// 执行密集计算
							for j := 0; j < 1000000; j++ {
								_ = j * j
							}
						}
					}()
				case "api_memory":
					// 为当前请求分配额外内存
					memory := make([]byte, level*1024*1024)
					defer func() { memory = nil }()
				case "api_db":
					// 模拟当前请求的数据库延迟
					time.Sleep(time.Duration(level) * time.Millisecond)
				}
			} else {
				// 系统级压力注入
				globalController.InjectPressure(pType, level, duration)
			}
		}
	}
	return "", nil
}
