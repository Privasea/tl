package tl

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type FeishuReq struct {
	MsgType string        `json:"msg_type"`
	Content FeishuContent `json:"content"`
}
type FeishuContent struct {
	Text string `json:"text"`
}
type FeishuText struct {
	AppName string `json:"app_name"`
	AppEnv  string `json:"app_env"`
	Msg     string `json:"msg"`
	Data    string `json:"data"`
	ErrMsg  string `json:"err_msg"`
}

// 发送通知到飞书
func SendToFeishu(feishuUrl string,msg string, data string, errMsg string) (err error) {
	//发送的内容
	feishuText := FeishuText{
		Msg:     msg,
		Data:    data,
		ErrMsg:  errMsg,
	}
	feishuTextData, err := json.Marshal(feishuText)
	if err != nil {
		return err
	}
	feishuReq := FeishuReq{
		MsgType: "text",
		Content: FeishuContent{
			Text: string(feishuTextData),
		},
	}
	// 将数据转换为JSON格式的bytes
	feishuReqData, err := json.Marshal(feishuReq)
	if err != nil {
		return err
	}
	// 创建HTTP请求
	req, err := http.NewRequest("POST", feishuUrl, bytes.NewBuffer(feishuReqData))
	if err != nil {
		return err
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 读取响应体
	/*_, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}*/
	return nil
}
