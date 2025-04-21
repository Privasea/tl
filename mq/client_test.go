package mq

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

var RespSuc = []byte(`{
    "data": {"code": "trace-http-test"},
    "message": "操作成功",
    "status_code": 200
}`)

type MQData struct {
	SessionNo  string    `json:"session_no"`
	Status     string    `json:"status"`
	FailureMsg string    `json:"failure_msg"`
	CreatedAt  time.Time `json:"created_at"`
	IsThird    bool      `json:"is_third"`
}

// 测试传入strut和slice和map和string和int
func TestClientMap(t *testing.T) {
	// 测试int的单条和多条
	err := send(int(1))
	if err != nil {
		t.Error(err)
	}

	err = sendMulti([]int{2, 3})
	if err != nil {
		t.Error(err)
	}
}

// 测试传入strut和slice和map和string和int
func TestClientMapPayment(t *testing.T) {
	var mqClient = NewMqClient(getNewGinContext(), "gmq",
		"https://api-dev-1.apisix.xthktech.cn",
		"GID_default_test3_payment",
		"https://paymentv3-local-dev-1.apisix.xthktech.cn/server-b/fast")

	mqClient.MqGroupID = "GID_default_test3_payment"

	var d = MQData{
		SessionNo:  "123123123",
		Status:     "succeed",
		FailureMsg: "",
		CreatedAt:  time.Now(),
		IsThird:    false,
	}

	err := mqClient.MqNormalMsg(d)
	if err != nil {
		log.Print(err.Error())
	}
}

func send(data interface{}) error {
	time := uint64(1000)
	orderKey := "OrderKey"
	var mqClient = NewMqClient(getNewGinContext(), "gmq",
		"https://api-dev-1.apisix.xthktech.cn",
		"GID_delay_dev1_rocketmq_test",
		"https://paymentv3-local-dev-1.apisix.xthktech.cn/server-b/fast")

	mqClient.MqGroupID = "GID_default_dev1_rocketmq_test"
	err := mqClient.MqNormalMsg(data)
	if err != nil {
		return err
	}

	mqClient.MqGroupID = "GID_order_dev1_rocketmq_test"
	err = mqClient.MqOrderMsg(orderKey, data)
	if err != nil {
		return err
	}

	mqClient.MqGroupID = "GID_delay_dev1_course_change"
	err = mqClient.MqFixEdTimerMsg(time, data)
	if err != nil {
		return err
	}

	var callback = MessageCallbackApi{
		CbAddress:   "https://paymentv3-local-dev-1.apisix.xthktech.cn/server-b/fast",
		CbMethod:    "POST",
		CbArguments: "xxxx",
		CbHeader: map[string]interface{}{
			"xxxx": "xxxx",
		},
	}
	mqClient.MqGroupID = "GID_transaction_dev1_rocketmq_test"
	err = mqClient.MqTransMsg(callback, data)
	if err != nil {
		return err
	}

	return nil
}

func sendMulti(data interface{}) error {
	time := uint64(1000)
	var mqClient = NewMqClient(getNewGinContext(), "gmq",
		"https://api-dev-1.apisix.xthktech.cn",
		"GID_delay_dev1_rocketmq_test",
		"https://paymentv3-local-dev-1.apisix.xthktech.cn/server-b/fast")

	mqClient.MqGroupID = "GID_default_dev1_rocketmq_test"
	err := mqClient.MqMultiNormalMsg(data)
	if err != nil {
		return err
	}

	mqClient.MqGroupID = "GID_delay_dev1_course_change"
	err = mqClient.MqMultiFixedTimerMsg(time, data)
	if err != nil {
		return err
	}

	return nil
}

func TestStartHttp(t *testing.T) {
	var serverMux = http.NewServeMux()
	serverMux.HandleFunc("/server-b/fast", func(w http.ResponseWriter, r *http.Request) {
		xx, _ := io.ReadAll(r.Body)
		log.Print(string(xx) + "\r\n")
		w.Write(RespSuc)
	})

	http.ListenAndServe("0.0.0.0:10080", serverMux)
}

func getNewGinContext() *gin.Context {
	ctx := new(gin.Context)
	uid := uuid.NewV4().String()
	ctx.Request = &http.Request{
		Header: make(map[string][]string),
	}
	ctx.Request.Header.Set("x-b3-traceid", uid)
	ctx.Request.Header.Set("x-huayu-traffic-tag", uid)
	ctx.Request.Header.Set("request_id", uid)
	ctx.Set("request_id", uid)
	ctx.Set("x-b3-traceid", uid)
	return ctx
}
