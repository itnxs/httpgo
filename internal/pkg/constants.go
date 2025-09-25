package pkg

import "time"

const (
	// 默认配置常量
	DefaultConnections  = 128
	DefaultDuration     = time.Second * 10
	DefaultTimeout      = time.Second * 3
	DefaultMaxRedirects = 30

	// HTTP 相关常量
	MIMEApplicationJSON = "application/json"
	MIMEApplicationForm = "application/x-www-form-urlencoded"

	// 统计相关常量
	StatisticsInterval = time.Millisecond * 10
	FieldWidth         = 18
	DefaultFps         = time.Duration(40)
	Padding            = 2
	MaxWidth           = 66
	ProcessColor       = "#444"

	// TUI 消息常量
	DoneMessage = 1

	// 颜色常量
	ColorGreen   = "#008700"
	ColorRed     = "#870000"
	ColorGray    = "#444"
	ColorYellow  = "#ffff00"
	ColorOrange  = "#ff8700"
	ColorDarkRed = "#870000"
)

// 协议相关常量
var (
	ProtocolHTTP  = []byte("http")
	ProtocolHTTPS = []byte("https")
)
