package mq

import (
	"github.com/satori/go.uuid"
)

// Message 消息体
type Message struct {
	MqOrder       string             `json:"mq_order"`       // 顺序消息使用，顺序消息Key。比如：一组消息需要顺序发送，这个字段就是这组消息的Key，其他消息该字段忽略
	MqDelayTime   uint64             `json:"mq_delaytime"`   // 延时消息使用，大于0表示是延时消息，单位毫秒，其他消息该字段忽略
	MqMessageId   string             `json:"mq_messageid"`   // 消息ID，唯一
	MqKey         string             `json:"mq_key"`         // Keys，rocketmq Key
	MQGroupID     string             `json:"mq_groupid"`     // mq groupid
	MqBody        string             `json:"mq_body"`        // 消息体
	MqCallbackApi MessageCallbackApi `json:"mq_callbackapi"` // 事务消息使用，发送方业务处理回调
}

// MessageCallbackApi 事务消息使用
type MessageCallbackApi struct {
	CbAddress   string                 `json:"cb_address"`   // 回调地址
	CbMethod    string                 `json:"cb_method"`    // 请求方式
	CbArguments string                 `json:"cb_arguments"` // MessageID业务调用自定义，用来识别回调的时候能知道是哪个消息
	CbHeader    map[string]interface{} `json:"cb_header"`    // 回调header参数
}

// MessageBody 消息body
type MessageBody struct {
	MqType       string      `json:"mq_type"`
	ApiAddress   string      `json:"api_address"`
	ApiMethod    string      `json:"api_method"`
	ApiArguments string      `json:"api_arguments"`
	ApiHeaders   []ApiHeader `json:"api_headers"`
}

// ApiHeader api header
type ApiHeader struct {
	Key string      `json:"key"`
	Val interface{} `json:"val"`
}

type NewMessageOptions func(m *Message) error

func NewMessage(opts ...NewMessageOptions) (*Message, error) {
	// 设置全局唯一key
	uuidKey := uuid.NewV4().String()
	message := &Message{
		MqOrder:       "",
		MqDelayTime:   0,
		MqMessageId:   uuidKey,
		MqKey:         uuidKey,
		MQGroupID:     "",
		MqBody:        "",
		MqCallbackApi: MessageCallbackApi{},
	}
	for _, opt := range opts {
		err := opt(message)
		if err != nil {
			return nil, err
		}
	}
	return message, nil
}

func SetGroupID(groupID string) NewMessageOptions {
	return func(m *Message) error {
		m.MQGroupID = groupID
		return nil
	}
}

func SetOrder(order string) NewMessageOptions {
	return func(m *Message) error {
		m.MqOrder = order
		return nil
	}
}

func SetDelayTime(delayTime uint64) NewMessageOptions {
	return func(m *Message) error {
		m.MqDelayTime = delayTime
		return nil
	}
}

func SetBody(body *MessageBody) NewMessageOptions {
	return func(m *Message) error {
		jsonBody, err := CJson.Marshal(body)
		if err != nil {
			return err
		}
		m.MqBody = string(jsonBody)
		return nil
	}
}

func SetCallbackApi(callback *MessageCallbackApi) NewMessageOptions {
	return func(m *Message) error {
		m.MqCallbackApi = *callback
		return nil
	}
}
