package logging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	timeFormat          = "15:04:05-02/01/06"
	logReportFileName   = "repeat_logs"
	crashReportFileName = "repeat_crashreport"
)

func NewLogger(version string, size int) Logger {
	rb := newRingBuffer(size)
	writer := logWriter(rb)
	return Logger{
		appVer: version,
		slog:   slog.New(slog.NewJSONHandler(writer, nil)),
		ring:   rb,
	}
}

type Logger struct {
	appVer string
	slog   *slog.Logger
	ring   *ringBuffer
}

func (l Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
	// NOTE: start goroutine to dump logs instead
	l.ring.SeenErr = true
}

func (l Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

func (l Logger) DumpLogs() {
	if !l.ring.SeenErr {
		return
	}
	now := time.Now()
	f, _ := os.Create(fmt.Sprintf("%v-%v.json", logReportFileName, now.Unix()))
	defer f.Close()

	fmt.Fprintf(f, "Version: %v\nOS: %v\nTime: %v\n\n", l.appVer, runtime.GOOS, now.Format(timeFormat))
	f.Write(l.ring.snapshot())
}

func (l Logger) dumpReport(ver string) {
	now := time.Now()
	f, _ := os.Create(fmt.Sprintf("%v-%v.json", crashReportFileName, now.Unix()))
	defer f.Close()

	report := struct {
		Version string
		OS      string
		Time    string
		Trace   []byte
	}{
		Version: ver,
		OS:      runtime.GOOS,
		Time:    now.Format(timeFormat),
		Trace:   l.ring.snapshot(),
	}
	encoder := json.NewEncoder(f)
	encoder.Encode(&report)
}

func (l Logger) DumpReport() {
	l.Error(string(debug.Stack()))
	l.dumpReport(l.appVer)
}
