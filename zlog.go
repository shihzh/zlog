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
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	innerLog, err := newLogger(&options)
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
	sink, err := AsyncLoggerSink(opt)
	if err != nil {
		return nil, err
	}

	getEncoder := func() zapcore.Encoder {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = epochFullTimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	encoder := getEncoder()

	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= opt.level
	})
	core := zapcore.NewCore(encoder, sink, levelEnabler)

	var zapLogger *zap.Logger
	if opt.disableCaller {
		zapLogger = zap.New(core)
	} else {
		zapLogger = zap.New(core, zap.AddCaller())
	}

	return zapLogger, nil
}
