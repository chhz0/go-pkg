package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// noopInfoLogger 是一个日志记录器的实现，它不对任何信息进行记录。
// 这种实现用于在不需要日志记录或者需要关闭日志记录的场景下。
// 它充当一个"空操作"的实现，即执行操作但不产生任何效果。
type noopInfoLogger struct{}

// Enabled implements InfoLogger.
func (n *noopInfoLogger) Enabled() bool { return false }

// Info implements InfoLogger.
func (n *noopInfoLogger) Info(ctx context.Context, msg string, args ...any) {}

// Infof implements InfoLogger.
func (n *noopInfoLogger) Infof(ctx context.Context, format string, args ...any) {}

// Infow implements InfoLogger.
func (n *noopInfoLogger) Infow(ctx context.Context, msg string, args ...any) {}

var disableInfoLogger = &noopInfoLogger{}

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug // -4
	LevelTrace = slog.Level(-2)
	LevelInfo  = slog.LevelInfo  //  0
	LevelWarn  = slog.LevelWarn  //  4
	LevelError = slog.LevelError //  8
)

var LevelIn = []Level{LevelDebug, LevelTrace, LevelInfo, LevelWarn, LevelError}

type logContextKey int

const (
	defaultLogContextKey logContextKey = iota
)

type SlogLogger struct {
	l   *slog.Logger
	lvl *slog.LevelVar
}

func New(level slog.Level) *SlogLogger {
	var lvl slog.LevelVar
	lvl.Set(level)

	sl := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,

		Level: &lvl,

		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 处理自定义的日志级别
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel := level.String()

				switch level {
				case LevelTrace:
					levelLabel = "TRACE"
				}
				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}))
	return &SlogLogger{l: sl, lvl: &lvl}
}

func (l *SlogLogger) SetLevel(level Level) {
	l.lvl.Set(level)
}

func (l *SlogLogger) Enabled() bool {
	return l.l.Enabled(context.Background(), l.lvl.Level())
}

func (l *SlogLogger) GetLogLevel() Level {
	var currentLevel Level = -10
	for _, level := range LevelIn {
		r := l.l.Enabled(context.Background(), level)
		if r {
			currentLevel = level
			break
		}
	}
	return currentLevel
}

func (l *SlogLogger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, defaultLogContextKey, l)
}

func (l *SlogLogger) clone() *SlogLogger {
	c := *l
	return &c
}

// log 是对 slog.Logger.log 的复制，
func (l *SlogLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.l.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	// NOTE: 这里修改 skip 为 4，*slog.Logger.log 源码中 skip 为 3
	runtime.Callers(4, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.l.Handler().Handle(ctx, r)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	// l.l.Info(msg, args...)
	l.Log(context.Background(), LevelInfo, msg, args...)
}

func (l *SlogLogger) Infof(format string, args ...any) {
	l.Log(context.Background(), LevelInfo, fmt.Sprintf(format, args...))
}

func (l *SlogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, LevelInfo, msg, args...)
}

func (l *SlogLogger) Trace(msg string, args ...any) {
	l.Log(context.Background(), LevelTrace, msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	// l.l.Warn(msg, args...)
	l.Log(context.Background(), LevelWarn, msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	// l.l.Error(msg, args...)
	l.Log(context.Background(), LevelError, msg, args...)
}

func (l *SlogLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}