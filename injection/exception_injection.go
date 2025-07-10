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

	// 使用通道来保持对内存块的引用
	memChan := make(chan []byte, memorySize)

	go func() {
		// 创建定时器用于监控内存使用情况
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		start := time.Now()
		allocated := 0

		for allocated < memorySize {
			// 分配1MB内存并写入数据以确保实际使用
			block := make([]byte, 1024*1024)
			for i := range block {
				block[i] = byte(i % 256)
			}

			// 将内存块放入通道，保持引用
			select {
			case memChan <- block:
				allocated++

				// 打印当前内存使用情况
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

			default:
				// 通道满了就等待一下
				time.Sleep(100 * time.Millisecond)
			}

			// 检查是否达到持续时间
			if time.Since(start) >= pressure.duration {
				break
			}
		}

		// 等待持续时间结束
		time.Sleep(pressure.duration)

		// 清理内存
		for len(memChan) > 0 {
			<-memChan
		}
		close(memChan)
		runtime.GC()

		// 更新状态
		pc.mu.Lock()
		if p, exists := pc.pressures[PressureMemory]; exists {
			p.active = false
		}
		pc.mu.Unlock()
	}()
}

// 修改DealPoints函数以支持压力注入
func DealPoints(server, path, method, imei, trackingPoints string) (string, error) {
	pointData, err := GetMatchPoints(server, path, method, imei, trackingPoints)
	//fmt.Println("测试一下钩子")
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

			// 系统级压力注入
			globalController.InjectPressure(pType, level, duration)
		}
	}
	return "", nil
}
