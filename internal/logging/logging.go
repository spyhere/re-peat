package logging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

const crashReportFileName = "repeat_crashreport"

func NewLogger(size int) Logger {
	return Logger{
		slog: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		ring: newRingBuffer(size),
	}
}

type Logger struct {
	slog *slog.Logger
	ring *ringBuffer
}

func (l Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
	l.ring.add(logEntry{
		Level: slog.LevelInfo,
		Msg:   msg,
	})
}

func (l Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
	l.ring.add(logEntry{
		Level: slog.LevelWarn,
		Msg:   msg,
	})
}

func (l Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
	l.ring.add(logEntry{
		Level: slog.LevelError,
		Msg:   msg,
	})
}

func (l Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
	l.ring.add(logEntry{
		Level: slog.LevelDebug,
		Msg:   msg,
	})
}

func (l Logger) DumpReport(ver string) {
	now := time.Now()
	f, _ := os.Create(fmt.Sprintf("%v:%v.json", crashReportFileName, now.Unix()))
	defer f.Close()

	report := struct {
		Version string
		OS      string
		Time    string
		Trace   []logEntry
	}{
		Version: ver,
		OS:      runtime.GOOS,
		Time:    now.Format("15:04:05-02/01/06"),
		Trace:   l.ring.snapshot(),
	}
	encoder := json.NewEncoder(f)
	encoder.Encode(&report)
}
