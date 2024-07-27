package logger

import (
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

type LevelLog = zapcore.Level

const (
	LevelDebug  = zapcore.DebugLevel
	LevelInfo   = zapcore.InfoLevel
	LevelWarn   = zapcore.WarnLevel
	LevelError  = zapcore.ErrorLevel
	LevelFatal  = zapcore.FatalLevel
	LevelDpanic = zapcore.DPanicLevel
	LevelPanic  = zapcore.PanicLevel
)

func InitLog(logDir string) {
	found, err := isDirExist(logDir)
	if err != nil {
		fmt.Println("installer log dir found error", err)
		os.Exit(1)
	}

	if !found {
		err := os.MkdirAll(logDir, common.FileMode0755)
		if err != nil {
			fmt.Println("create log dir error", err)
			os.Exit(1)
		}
	}

	p := path.Join(logDir, "terminus_install.log")
	file, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, common.FileMode0755)
	if err != nil {
		panic(err)
	}

	consolePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel || lvl == zapcore.ErrorLevel
	})
	filePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	fileEncoder := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey: "T",
		// NameKey:        "N",
		// CallerKey:      "C",
		// FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleDebugging := zapcore.Lock(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoderConfig), consoleDebugging, consolePriority),
		zapcore.NewCore(zapcore.NewJSONEncoder(fileEncoder), zapcore.AddSync(file), filePriority),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.FatalLevel)).Sugar()
	defer logger.Sync()
}

func isDirExist(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func getLevel(level string) (l zapcore.Level) {
	switch level {
	case "debug":
		l = zap.DebugLevel
	case "info":
		l = zap.InfoLevel
	case "warn":
		l = zap.WarnLevel
	case "error":
		l = zap.ErrorLevel
	case "panic":
		l = zap.PanicLevel
	case "fatal":
		l = zap.FatalLevel
	default:
		l = zap.InfoLevel
	}
	return
}

func Sync() error {
	return logger.Sync()
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Debugw(msg string, args ...any) {
	logger.Debugw(msg, args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Infow(msg string, args ...any) {
	logger.Infow(msg, args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Warnw(msg string, args ...any) {
	logger.Warnw(msg, args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Errorw(msg string, args ...any) {
	logger.Errorw(msg, args...)
}

func DPanic(args ...any) {
	logger.DPanic(args...)
}

func DPanicf(format string, args ...any) {
	logger.DPanicf(format, args...)
}

func DPanicw(msg string, args ...any) {
	logger.DPanicw(msg, args...)
}

func Panic(args ...any) {
	logger.Panic(args...)
}

func Panicf(format string, args ...any) {
	logger.Panicf(format, args...)
}

func Panicw(msg string, args ...any) {
	logger.Panicw(msg, args...)
}

func Fatal(args ...any) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Fatalw(msg string, args ...any) {
	logger.Fatalw(msg, args...)
}

func GetLogger() *zap.SugaredLogger {
	return logger
}