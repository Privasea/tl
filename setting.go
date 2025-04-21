package tl

import (
	"github.com/Privasea/tl/dbx"
	"github.com/Privasea/tl/gcalx"
	"github.com/Privasea/tl/injection"
	"github.com/Privasea/tl/logx"
	"github.com/Privasea/tl/mq"
)

// 必须设置项
var SetLogger = logx.SetLogger
var SetDbLog = logx.SetDbLog
var SetLogErr = injection.SetLogErr
var SetParamLog = setParamLog
var SetSdkLog = gcalx.SetSdkLog

// 日志打印项
var InfoF = logx.InfoF
var WarnF = logx.WarnF
var ErrorF = logx.ErrorF

// 请求包
// var Begin = gcalx.DefaultClient().Begin
var Get = gcalx.DefaultClient().Get
var Post = gcalx.DefaultClient().Post
var PostJson = gcalx.DefaultClient().PostJson
var WithOption = gcalx.DefaultClient().WithOption
var WithOptions = gcalx.DefaultClient().WithOptions
var WithHeader = gcalx.DefaultClient().WithHeader
var WithHeaders = gcalx.DefaultClient().WithHeaders
var WithContext = gcalx.DefaultClient().WithITrace

// 数据库
var InitConn = dbx.InitConn
var WarpMysql = dbx.Wrap

// mq
var NewMqClient = mq.NewMqClient

type MessageCallbackApi = mq.MessageCallbackApi

const (
	PROXY_HTTP int = iota
	PROXY_SOCKS4
	PROXY_SOCKS5
	PROXY_SOCKS4A

	// CURL like OPT
	OPT_AUTOREFERER
	OPT_FOLLOWLOCATION
	OPT_CONNECTTIMEOUT
	OPT_CONNECTTIMEOUT_MS
	OPT_MAXREDIRS
	OPT_PROXYTYPE
	OPT_TIMEOUT
	OPT_TIMEOUT_MS
	OPT_COOKIEJAR
	OPT_INTERFACE
	OPT_PROXY
	OPT_REFERER
	OPT_USERAGENT

	// Other OPT
	OPT_REDIRECT_POLICY
	OPT_PROXY_FUNC
	OPT_DEBUG
	OPT_UNSAFE_TLS

	OPT_CONTEXT
)
