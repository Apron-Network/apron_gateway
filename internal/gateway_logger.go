package internal

import (
	"os"
	"path/filepath"
)

const logBufferSize = 100 * 1024

type GatewayLogger struct {
	LogFile    string
	msgChannel chan string
	logFP      *os.File
}

func (l *GatewayLogger) Init() {
	l.initLogFile()
	l.msgChannel = make(chan string, logBufferSize)

	go func() {
		for msg := range l.msgChannel {
			l.logFP.WriteString(msg)
		}
	}()
}

func (l *GatewayLogger) RecordMsg(msg string) {
	l.msgChannel <- msg
}

func (l *GatewayLogger) initLogFile() {
	absLogPath, err := filepath.Abs(l.LogFile)
	CheckError(err)

	basePath := filepath.Base(absLogPath)
	CheckError(err)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		os.MkdirAll(basePath, 0755)
	}

	l.logFP, err = os.OpenFile(absLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)
}
