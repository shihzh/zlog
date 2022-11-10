package zlog

import (
	"testing"
	"time"
)

func TestDefaultLog(t *testing.T) {
	log, err := NewLogger()
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	logger := log.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))

	logger.Sync()
}

func TestBigMsg(t *testing.T) {
	logger := GetDefaultLogger().Sugar()

	v := make([]byte, 0, 8000)
	for i := 0; i < 8000; i++ {
		v = append(v, '1')
	}
	logger.Infof("hello %s", string(v))

	logger.Sync()
}

func Test2Logger2file(t *testing.T) {
	log, err := NewLogger(
		DisableCaller(false),
		LogPath("log/withCaller.txt"))
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	log1, err := NewLogger(
		DisableCaller(true),
		LogPath("log/withoutCaller.txt"))
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	logger := log.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))

	logger = log1.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))

	log.Sync()
	log1.Sync()
}

func Test2Logger1file(t *testing.T) {
	log, err := NewLogger(
		DisableCaller(false),
		LogPath("log/a.txt"))
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	log1, err := NewLogger(
		DisableCaller(true),
		LogPath("log/a.txt"))
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	logger := log.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))

	logger = log1.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))

	logger.Sync()
}

func TestStdout(t *testing.T) {
	log, err := NewLogger(Stdout(true))
	if err != nil {
		t.Errorf("InitLog failed, %s", err.Error())
		return
	}

	logger := log.Sugar()
	logger.Infof("hello %s", string("Infof"))
	logger.Warnf("hello %s", string("Warnf"))
	logger.Errorf("hello %s", string("Errorf"))
}

func BenchmarkLog(b *testing.B) { // 4281 ns/op
	logger := GetDefaultLogger().Sugar()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		logger.Info("hello ", string("Infof"), i)
	}
	b.Logf("%d ops,  %s", b.N, time.Since(start))
}

func BenchmarkLogf(b *testing.B) { // 4318 ns/op
	logger := GetDefaultLogger().Sugar()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		logger.Infof("hello %s, %d", string("Infof"), i)
	}
	b.Logf("%d ops,  %s", b.N, time.Since(start))
}
