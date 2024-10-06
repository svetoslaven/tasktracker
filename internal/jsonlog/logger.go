package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelDiscard
)

func (lvl Level) String() string {
	switch lvl {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	writer    io.Writer
	minLvl    Level
	writeLock sync.Mutex
}

func NewLogger(w io.Writer, minLvl Level) *Logger {
	return &Logger{
		writer: w,
		minLvl: minLvl,
	}
}

var DiscardLogger = &Logger{
	writer: io.Discard,
	minLvl: LevelDiscard,
}

func (l *Logger) Write(msg []byte) (n int, err error) {
	return l.log(LevelError, string(msg), nil)
}

func (l *Logger) LogInfo(msg string, properties map[string]string) {
	l.log(LevelInfo, msg, properties)
}

func (l *Logger) LogError(err error, properties map[string]string) {
	l.log(LevelError, err.Error(), properties)
}

func (l *Logger) LogFatal(err error, properties map[string]string) {
	l.log(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

func (l *Logger) log(lvl Level, msg string, properties map[string]string) (int, error) {
	if lvl < l.minLvl {
		return 0, nil
	}

	entry := struct {
		Lvl        string            `json:"level"`
		Time       string            `json:"time"`
		Msg        string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Lvl:        lvl.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Msg:        msg,
		Properties: properties,
	}

	if lvl >= LevelError {
		entry.Trace = string(debug.Stack())
	}

	var line []byte

	line, err := json.Marshal(entry)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	l.writeLock.Lock()
	defer l.writeLock.Unlock()

	return l.writer.Write(append(line, '\n'))
}
