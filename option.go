package zlog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

const (
	defaultLogPath = "./log/%s.log"
	maxFileSize    = 256 // 256MBytes
	maxBackups     = 100 // 最多保留日志文件个数
	maxAge         = 7   // 最多保留日志文件天数
)

// Options 属性
type Options struct {
	level         zapcore.Level          // 测试环境日志级别为debug
	logPath       string                 // 日志路径
	withGID       bool                   // 打印协程id
	stdout        bool                   // 日志同时打印到标准输出
	overflow      bool                   // 日志缓存管道溢出则丢弃日志
	disableCaller bool                   // 打印调用位置
	rotate        bool                   // 是否使用lumberjack滚动日志
	compress      bool                   // 日志文件是否压缩
	maxFileSize   int                    // 日志文件分割大小，单位MByte
	maxBackups    int                    // 日志文件保留个天数
	maxAge        int                    // 日志文件保留天数
	bufioSize     int                    // 写文件io的缓存大小
	fields        map[string]interface{} // 日志默认附加的字段
}

// 默认属性
var defaultOptions = Options{
	level:         zap.InfoLevel,
	withGID:       false,
	stdout:        false,
	overflow:      false,
	disableCaller: false,
	rotate:        true,
	compress:      true,
	maxFileSize:   maxFileSize,
	maxBackups:    maxBackups,
	maxAge:        maxAge,
	bufioSize:     256 * 1024,
}

// 默认文件路径：./log/进程名.log
func getLogFilePath(opt *Options) string {
	if len(opt.logPath) == 0 {
		return fmt.Sprintf(defaultLogPath, processName())
	}
	return opt.logPath
}

// 获取进程名
func processName() string {
	processPath := os.Args[0]
	if runtime.GOOS == "windows" {
		processPath = filepath.ToSlash(processPath)
	}
	return path.Base(processPath)
}

// Option 属性选项
type Option func(*Options)

// WithLevel 设置日志级别
func WithLevel(level zapcore.Level) Option {
	return func(o *Options) {
		o.level = level
	}
}

// Overflow 设置日志缓存管道溢出后是否丢弃
func Overflow(discard bool) Option {
	return func(o *Options) {
		o.overflow = discard
	}
}

// WithGID 打印协程ID
func WithGID(withGID bool) Option {
	return func(o *Options) {
		o.withGID = withGID
	}
}

// BufioSize bufio缓存的大小, 默认1024*8
func BufioSize(bufioSize int) Option {
	return func(o *Options) {
		o.bufioSize = bufioSize
	}
}

// LogPath 日志文件路径
func LogPath(logPath string) Option {
	return func(o *Options) {
		o.logPath = logPath
	}
}

// DisableCaller 调用位置，非常影响性能
func DisableCaller(disableCaller bool) Option {
	return func(o *Options) {
		o.disableCaller = disableCaller
	}
}

// Rotate 滚动日志
func Rotate(rotate bool) Option {
	return func(o *Options) {
		o.rotate = rotate
	}
}

// RotateOpt 滚动日志选项
func RotateOpt(compress bool, maxFileSize, maxBackups, maxAge int) Option {
	return func(o *Options) {
		o.rotate = true
		o.compress = compress
		o.maxFileSize = maxFileSize
		o.maxBackups = maxBackups
		o.maxAge = maxAge
	}
}

// WithFields 所有日志都附带的字段
func WithFields(fields map[string]interface{}) Option {
	return func(o *Options) {
		o.fields = fields
	}
}

// Stdout 日志打印到标准输出
func Stdout(stdout bool) Option {
	return func(o *Options) {
		o.stdout = stdout
	}
}
