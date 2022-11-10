package zlog

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	appInnerLog *zap.Logger
)

func init() {
	// 将AsyncLoggerSink工厂函数注册到zap中, 自定义协议名为 AsyncLog
	_ = zap.RegisterSink("AsyncLog", AsyncLoggerSink)
	loggerList.opts = make(map[string]*Options)
	loggerList.sink = make(map[string]zap.Sink)
}

// GetDefaultLogger 获取默认日志，文件名=log/进程名.log
func GetDefaultLogger() *zap.Logger {
	if appInnerLog != nil {
		return appInnerLog
	}

	var err error
	appInnerLog, err = NewLogger()
	if err != nil {
		fmt.Printf("InitLog failed with %s\n", err.Error())
		return nil
	}

	return appInnerLog
}

// NewLogger 初始化日志
func NewLogger(opts ...Option) (*zap.Logger, error) {
	for _, opt := range opts {
		opt(&defaultOptions)
	}

	innerLog, err := newLogger(&defaultOptions)
	if err != nil {
		return nil, err
	}
	return innerLog, nil
}

func epochFullTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// newLogger 初始化日志
func newLogger(opt *Options) (*zap.Logger, error) {
	var url string
	if opt.logPath != "" {
		url = fmt.Sprint("AsyncLog://127.0.0.1/", opt.logPath)
	} else {
		url = fmt.Sprint("AsyncLog://127.0.0.1")
	}

	outPaths := []string{url}
	if opt.stdout {
		outPaths = append(outPaths, "stdout")
	}

	loggerList.Lock()
	loggerList.opts[url] = opt
	loggerList.Unlock()

	defer func() {
		loggerList.Lock()
		delete(loggerList.opts, url)
		loggerList.Unlock()
	}()

	encoder := func() zapcore.EncoderConfig {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = epochFullTimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return encoderConfig
	}()

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(opt.level)
	config.Encoding = "console"
	config.EncoderConfig = encoder
	config.OutputPaths = outPaths
	config.DisableCaller = opt.disableCaller
	log, err := config.Build(zap.AddStacktrace(zap.ErrorLevel) /*zap.AddCaller()*/)
	if err != nil {
		return nil, err
	}
	return log, nil
}
