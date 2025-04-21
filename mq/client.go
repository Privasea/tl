package mq

import (
	"facebyte/pkg/tl/gcalx"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"net/url"
	"reflect"
	"time"
)

type Client struct {
	Ctx          *gin.Context  `json:"ctx"`           // ctx
	Domain       string        `json:"domain"`        // 服务地址
	ClientHeader ClientHeader  `json:"client_header"` // 请求header
	Path         string        `json:"path"`          // 接口地址
	Timeout      time.Duration `json:"timeout"`       // 超时时间
	ServiceName  string        `json:"service_name"`  // 服务名称
	TargetAddr   string        `json:"target_addr"`   // 目标端地址
	MqGroupID    string        `json:"mq_group_id"`   // 使用mq的group_id
}

func NewMqClient(ctx *gin.Context, appName, domain, groupID, addr string) *Client {
	client := &Client{
		Ctx: ctx,
		ClientHeader: map[string]string{
			"Target-Service-Name": "common-mqpub-pro",
		},
		ServiceName: appName,
		Domain:      domain,
		Timeout:     time.Second * 5,
		TargetAddr:  addr,
		MqGroupID:   groupID,
	}

	return client
}

type ClientHeader map[string]string

type MqServiceResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	RequestId  string      `json:"request_id"`
	Data       interface{} `json:"data"`
}

func (client *Client) httpSend(message interface{}) (*MqServiceResponse, error) {
	// 验证ctx是否传入
	if client.Ctx == nil {
		return nil, errors.New("ctx需要传递")
	}

	url1 := client.Domain + client.Path
	httpClient := gcalx.DefaultClient().WithITrace(client.Ctx).
		WithHeader("content-type", "application/json").
		WithOption(gcalx.OPT_TIMEOUT, int(client.Timeout.Seconds()))
	for key, value := range client.ClientHeader {
		httpClient.WithHeader(key, value)
	}

	res, err := httpClient.PostJson(url1, message)
	if err != nil {
		return nil, err
	}
	body, err := res.ReadAll()
	if err != nil {
		return nil, err
	}
	var mqServiceResponse MqServiceResponse
	err = CJson.Unmarshal(body, &mqServiceResponse)
	if mqServiceResponse.StatusCode != 200 {
		return nil, errors.New(mqServiceResponse.Message)
	}
	return &mqServiceResponse, err
}

func (client *Client) parseMessage(data interface{}) (*Message, error) {
	traffic := client.Ctx.GetString("x-huayu-traffic-tag")
	if traffic == "" && client.Ctx.GetString("is_cross_middleware") == "" {
		traffic = client.Ctx.GetHeader("x-huayu-traffic-tag")
	}
	requestId := client.Ctx.GetString("trace_id")
	if requestId == "" && client.Ctx.GetString("is_cross_middleware") == "" {
		requestId = client.Ctx.GetHeader("trace_id")
	}
	if requestId == "" { //如果没有传入trace_id，则生成request_id
		requestId = uuid.NewV4().String()
	}
	var apiHeaders = []ApiHeader{
		{Key: "content-type", Val: "application/json"},
		{Key: "x-huayu-traffic-tag", Val: traffic},
		{Key: "request_id", Val: requestId},
	}

	msg, err := CJson.Marshal(&data)
	if err != nil {
		return nil, err
	}

	apiAddress, err := url.QueryUnescape(client.TargetAddr)
	if err != nil {
		return nil, err
	}

	messageBody := &MessageBody{
		MqType:       "http",
		ApiAddress:   apiAddress,
		ApiMethod:    "POST",
		ApiArguments: string(msg),
		ApiHeaders:   apiHeaders,
	}
	return NewMessage(
		SetGroupID(client.MqGroupID),
		SetBody(messageBody),
	)
}

func (client *Client) parseMessages(data []interface{}) ([]*Message, error) {
	var msgs []*Message
	for _, v := range data {
		var apiHeaders = []ApiHeader{{Key: "content-type", Val: "application/json"}}

		msg, err := CJson.Marshal(&v)
		if err != nil {
			return nil, err
		}

		apiAddress, err := url.QueryUnescape(client.TargetAddr)
		if err != nil {
			return nil, err
		}

		messageBody := &MessageBody{
			MqType:       "http",
			ApiAddress:   apiAddress,
			ApiMethod:    "POST",
			ApiArguments: string(msg),
			ApiHeaders:   apiHeaders,
		}
		m, err := NewMessage(
			SetGroupID(client.MqGroupID),
			SetBody(messageBody),
		)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// MqNormalMsg 普通消息投递
func (client *Client) MqNormalMsg(data interface{}, msgKey ...string) error {
	client.Path = "/api/normalmsg"

	message, err := client.parseMessage(data)
	if err != nil {
		return err
	}
	if len(msgKey) > 0 {
		message.MqKey = msgKey[0]
	}

	_, err = client.httpSend(message)
	if err != nil {
		return err
	}

	return nil
}

// MqOrderMsg 顺序消息
func (client *Client) MqOrderMsg(orderKey string, data interface{}, msgKey ...string) error {
	client.Path = "/api/ordermsg"

	message, err := client.parseMessage(data)
	if err != nil {
		return err
	}
	message.MqOrder = orderKey
	if len(msgKey) > 0 {
		message.MqKey = msgKey[0]
	}
	_, err = client.httpSend(message)
	if err != nil {
		return err
	}
	return nil
}

// MqFixEdTimerMsg 延时消息
func (client *Client) MqFixEdTimerMsg(delayTime uint64, data interface{}, msgKey ...string) error {
	client.Path = "/api/fixedtimermsg"
	message, err := client.parseMessage(data)
	if err != nil {
		return err
	}
	if len(msgKey) > 0 {
		message.MqKey = msgKey[0]
	}
	message.MqDelayTime = delayTime
	_, err = client.httpSend(message)
	if err != nil {
		return err
	}
	return nil
}

// MqTransMsg 事务消息
func (client *Client) MqTransMsg(callBack MessageCallbackApi, data interface{}, msgKey ...string) error {
	client.Path = "/api/transmsg"

	message, err := client.parseMessage(data)
	if err != nil {
		return err
	}

	if len(msgKey) > 0 {
		message.MqKey = msgKey[0]
	}

	message.MqCallbackApi = callBack
	_, err = client.httpSend(message)
	if err != nil {
		return err
	}
	return nil
}

// MqMultiNormalMsg 多条普通消息投递
func (client *Client) MqMultiNormalMsg(data interface{}, msgKey ...string) error {
	client.Path = "/api/multi-normalmsg"
	var datas []interface{}

	v := reflect.ValueOf(data)
	for i := 0; i < v.Len(); i++ {
		ele := v.Index(i)
		datas = append(datas, ele.Interface())
	}

	if len(datas) > 0 {
		messages, err := client.parseMessages(datas)
		if err != nil {
			return err
		}
		if len(msgKey) > 0 {
			for key, _ := range messages {
				messages[key].MqKey = msgKey[0]
			}
		}
		_, err = client.httpSend(messages)
		if err != nil {
			return err
		}
	}

	return nil
}

// MqMultiFixedTimerMsg 多条定时消息
func (client *Client) MqMultiFixedTimerMsg(delayTime uint64, data interface{}) error {
	client.Path = "/api/multi-fixedtimermsg"
	var datas []interface{}

	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Slice {
		return errors.New("只接受slice数据")
	}

	v := reflect.ValueOf(data)
	for i := 0; i < v.Len(); i++ {
		ele := v.Index(i)
		datas = append(datas, ele.Interface())
	}

	if len(datas) > 0 {
		messages, err := client.parseMessages(datas)
		if err != nil {
			return err
		}

		for _, m := range messages {
			m.MqDelayTime = delayTime
		}

		_, err = client.httpSend(messages)
		if err != nil {
			return err
		}
	}

	return nil
}
