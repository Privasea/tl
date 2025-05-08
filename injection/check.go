package injection

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// RequestData 定义了请求数据的结构体
type GetApiDataForInjectRequestData struct {
	Server string `json:"server"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Imei   string `json:"imei,omitempty"` // 使用omitempty，表示如果 imei 为空，不会包含在 JSON 中
}

// Response 定义了响应的结构体
type GetApiDataForInjectResponse struct {
	Code int    `json:"code"` // 状态码
	Data Data   `json:"data"` // 数据
	Msg  string `json:"msg"`  // 错误消息或成功消息
}
type Data struct {
	List []LogItem `json:"list"`
}

// Data 定义了 data 数组中的每一项结构体
type LogItem struct {
	ID            int    `json:"id"`             // 异常 ID
	Server        string `json:"server"`         // 服务名
	Path          string `json:"path"`           // 接口路径
	Method        string `json:"method"`         // 接口方法
	Imei          string `json:"imei"`           // IMEI 号码
	PointCode     int    `json:"point_code"`     // 点位编号
	PointType     string `json:"point_type"`     // 点位类型
	PointDesc     string `json:"point_desc"`     // 点位描述
	InjectionType string `json:"injection_type"` // 异常注入类型
	InjectionData string `json:"injection_data"` // 异常注入数据
}

func GetApiDataForInject(server, path, method, imei string) ([]LogItem, error) {
	if dl {
		//client := NewClient("http://34.13.77.118/api/v1/inner_use")
		client := NewClient("http://exception-service:5000/api/v1/inner_use")

		req := GetApiDataForInjectRequestData{
			Server: server,
			Path:   path,
			Method: method,
			Imei:   imei,
		}

		postResponse, err := client.Post("/exception/check", req)
		if err != nil {
			return nil, err
		}

		// 解析 POST 请求的响应数据
		var response GetApiDataForInjectResponse
		err = json.Unmarshal(postResponse, &response)
		if err != nil {
			return nil, err
		}
		if response.Code != 0 {
			return nil, errors.New(response.Msg)
		}
		return response.Data.List, nil
	}
	//var mockList []LogItem
	//mock := LogItem{
	//	ID:            1,
	//	Server:        "facebyte",
	//	Path:          "/v1/sys/config",
	//	Method:        "GET",
	//	Imei:          "1111",
	//	PointCode:     1002,
	//	PointType:     "dblog",
	//	PointDesc:     "",
	//	InjectionType: "err",
	//	InjectionData: "ErrInvalidTransaction",
	//}
	//mockList = append(mockList, mock)
	//return mockList, nil
	return nil, nil
}
