package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	timeFormat          = "15:04:05-02/01/06"
	LogReportFileName   = "repeat_logs"
	CrashReportFileName = "repeat_crashreport"
)

func NewLogger(version string, size int) Logger {
	rb := newRingBuffer(size)
	writer := logWriter(rb)
	dumpDone := make(chan struct{})
	l := Logger{
		appVer:     version,
		slog:       slog.New(slog.NewJSONHandler(writer, nil)),
		ring:       rb,
		dumpCh:     make(chan struct{}),
		DumpDoneCh: dumpDone,
	}
	go func() {
		for range l.dumpCh {
			l.dumpFile(LogReportFileName)
			dumpDone <- struct{}{}
		}
	}()
	return l
}

type Logger struct {
	appVer     string
	dumpCh     chan struct{}
	DumpDoneCh chan struct{}
	slog       *slog.Logger
	ring       *ringBuffer
}

func (l Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l Logger) Error(msg string, err error) {
	l.slog.Error(msg, "err", err)
	select {
	case l.dumpCh <- struct{}{}:
	default:
	}
}

func (l Logger) Crash(args ...any) {
	l.slog.Error("CRASH", args...)
	l.dumpFile(CrashReportFileName)
}

func (l Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// FIX: log errors to tmp/log
func (l Logger) dumpFile(filePrefix string) {
	now := time.Now()
	filename := fmt.Sprintf("%v-%v.txt", filePrefix, now.Unix())
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	f, _ := os.Create(filepath.Join(desktop, filename))
	defer f.Close()

	fmt.Fprintf(f, "Version: %v\nOS: %v\nTime: %v\n\n", l.appVer, runtime.GOOS, now.Format(timeFormat))
	f.Write(l.ring.snapshot())
	if filePrefix == CrashReportFileName {
		f.WriteString("\n")
		f.Write(debug.Stack())
	}
}
