package zlog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxChanSize = 256 * 1024 // 256个管道，每个管道容量为1024个[]byte，共256*1024个切片消息
)

var (
	loggerList loggerMap
)

// Flusher .
type Flusher interface {
	Flush() error
}

// WriteCloseFlusher .
type WriteCloseFlusher struct {
	io.Writer
	io.Closer
	Flusher
}

// AsyncLogSink 定义一个结构体
type AsyncLogSink struct {
	closed     bool
	stdout     bool
	failCounts uint64
	filepath   string
	chanMgr    *ChanMgr
	writer     *WriteCloseFlusher
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

type loggerMap struct {
	sync.Mutex
	sink map[string]zap.Sink
}

func AsyncLoggerSink(opts *Options) (sink zap.Sink, err error) {
	var writer io.WriteCloser

	loggerList.Lock()
	defer loggerList.Unlock()

	v, ok := loggerList.sink[opts.logPath]
	if ok {
		return v, nil
	}

	filePath := opts.logPath
	if filePath == "" {
		filePath = getLogFilePath(&defaultOptions)
	}

	if opts.rotate {
		writer = &lumberjack.Logger{
			Filename:   filePath,
			Compress:   opts.compress,
			LocalTime:  true,
			MaxSize:    opts.maxFileSize,
			MaxBackups: opts.maxBackups,
			MaxAge:     opts.maxAge,
		}
	} else {
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return nil, err
		}
		// 使用第三方程序rotate的话，用append模式打开，否则会形成空洞的大文件
		openFlag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
		writer, err = os.OpenFile(filePath, openFlag, os.FileMode(0644))
		if err != nil {
			return nil, err
		}
	}

	bw := bufio.NewWriterSize(writer, opts.bufioSize)
	wc := &WriteCloseFlusher{
		Writer:  bw,
		Flusher: bw,
		Closer:  writer,
	}
	c := &AsyncLogSink{
		stdout:   opts.stdout,
		filepath: filePath,
		writer:   wc,
		chanMgr:  NewChanMgr(256, maxChanSize/256),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.wg.Add(1)
	go func() {
		c.loop()
		c.wg.Done()
	}()

	loggerList.sink[opts.logPath] = c
	return c, nil
}

// Sync 因文件io只可AsyncLogSink::loop协程中使用，否则引发同步问题
func (c *AsyncLogSink) Sync() error {
	return nil
}

// Close 定义Close方法以实现Sink接口
func (c *AsyncLogSink) Close() error {
	time.Sleep(time.Millisecond * 2) // 短暂等待日志写入管道

	if c.closed {
		return nil
	}
	c.closed = true
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait() // wait until all msgs have been consumed
	c.writer.Close()

	loggerList.Lock()
	defer loggerList.Unlock()
	delete(loggerList.sink, c.filepath)

	return nil
}

// 定义Write方法以实现Sink接口
func (c *AsyncLogSink) Write(p []byte) (n int, err error) {
	// zap框架复用切片p参数,需要拷贝否则错乱
	cp := make([]byte, len(p))
	copy(cp, p)

	msgChan, _ := c.chanMgr.NextWrite()
	if !defaultOptions.overflow {
		msgChan <- cp
	} else {
		select {
		case msgChan <- cp:
		default:
			failCounts := atomic.AddUint64(&c.failCounts, 1)
			if failCounts%100 == 0 {
				msgChan <- addField(failCounts, "blockNums", cp)
			}
		}
	}
	return len(p), nil
}

func addField(failCounts uint64, name string, msg []byte) []byte {
	b := bytes.TrimSuffix(msg, []byte("}\n"))
	b = append(b, []byte(fmt.Sprintf(",\"%s\":%d}\n", name, failCounts))...)
	return b
}

func (c *AsyncLogSink) loop() {
	defer func() {
		recover()
	}()

	var msg []byte
	closed := false

	for {
		msgChan, idx := c.chanMgr.NextRead()
		if !closed {
			select {
			case msg = <-msgChan:
			case <-c.ctx.Done():
				closed = true
			}
		} else {
			select {
			case msg = <-msgChan:
			default:
				c.writer.Flush()
				return
			}
		}

		if len(msg) > 0 {
			c.writer.Write(msg)
			if c.stdout {
				os.Stdout.Write(msg)
			}
			msg = nil
		}

		if c.chanMgr.Len(idx+1) == 0 {
			c.writer.Flush()
		}
	}
}
