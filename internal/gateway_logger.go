package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const logBufferSize = 100 * 1024 // Default buf size 100KB

type LoggingLevels int

type GatewayLogger struct {
	LogFile   string
	MaxRotate int

	msgChannel  chan string
	logFile     *os.File
	writer      *bufio.Writer
	writtenSize int
	logLock     sync.Mutex
	logfileName string
	logfileBase string
	logfileExt  string
	logfileDir  string
	lastLogTime time.Time
}

func (l *GatewayLogger) Init() {
	l.initLogFile()
	l.msgChannel = make(chan string, logBufferSize)

	go func() {
		for msg := range l.msgChannel {
			fmt.Printf("Received message from channel: %s\n", msg)
			bytesWritten, err := l.writer.WriteString(msg)
			CheckError(err)

			l.writtenSize += bytesWritten
			l.lastLogTime = time.Now()

		}
	}()

	// Flush buffer writer every minute, and check whether rotations required
	go func() {
		for {
			time.Sleep(time.Minute)
			l.writer.Flush()

			// Rotate file every day at midnight
			if l.rotationRequired() {
				err := l.rotateFile()
				CheckError(err)
			}
		}
	}()
}

func (l *GatewayLogger) RecordMsg(msg string) {
	l.msgChannel <- msg
}

func (l *GatewayLogger) rotateFile() error {
	l.logLock.Lock()
	defer l.logLock.Unlock()

	var err error

	if l.writer != nil {
		if err = l.writer.Flush(); err != nil {
			return err
		}
	}

	if l.logFile != nil {
		if err = l.logFile.Close(); err != nil {
			return err
		}
		l.logFile = nil
	}

	_, err = os.Stat(l.LogFile)
	if err == nil && l.writtenSize != 0 {
		// log file existing and written size is not 0, need rotation
		rotatedLogfileName := fmt.Sprintf("%s.%s%s",
			l.logfileBase,
			time.Now().UTC().Format("20060102150405"),
			l.logfileExt)
		rotatedLogfilePath := filepath.Join(l.logfileDir, rotatedLogfileName)
		fmt.Println("Old file rotated to: ", rotatedLogfilePath)
		if err = os.Rename(l.LogFile, rotatedLogfilePath); err != nil {
			return err
		}

		// Tidy up log dir to remove older rotated file
		go func() {
			logFilenamePattern := fmt.Sprintf("%s.*%s", l.logfileBase, l.logfileExt)
			globPattern := filepath.Join(l.logfileDir, logFilenamePattern)
			fmt.Printf("Glob pattern: %s\n", globPattern)
			rotatedFiles, _ := filepath.Glob(globPattern)
			shouldBeRemovedCount := len(rotatedFiles) - l.MaxRotate
			fmt.Printf("Rotated file count: %d, max size: %d, should be removed: %d\n", len(rotatedFiles), l.MaxRotate, shouldBeRemovedCount)
			if shouldBeRemovedCount > 0 {
				sort.Strings(rotatedFiles)
				for _, fileName := range rotatedFiles[:shouldBeRemovedCount] {
					err = os.Remove(fileName)
					CheckError(err)
				}
			}
		}()
	}

	l.initLogWriter()
	return nil
}

func (l *GatewayLogger) initLogFile() {
	// Create log path if not existing
	absLogPath, err := filepath.Abs(l.LogFile)
	CheckError(err)

	l.logfileDir = filepath.Dir(absLogPath)
	CheckError(err)

	if _, err := os.Stat(l.logfileDir); os.IsNotExist(err) {
		os.MkdirAll(l.logfileDir, 0755)
	}

	// Set default value for rotate count
	if l.MaxRotate == 0 {
		l.MaxRotate = 5
	}

	l.initLogWriter()

	_, l.logfileName = filepath.Split(l.LogFile)
	l.logfileExt = filepath.Ext(l.LogFile)
	l.logfileBase = l.logfileName[:len(l.logfileName)-len(l.logfileExt)]
}

func (l *GatewayLogger) initLogWriter() {
	absLogPath, err := filepath.Abs(l.LogFile)
	l.logFile, err = os.OpenFile(absLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckError(err)

	l.writer = bufio.NewWriterSize(l.logFile, logBufferSize)
	l.writtenSize = 0
}

//rotationRequired checks whether logging requires rotation
func (l *GatewayLogger) rotationRequired() bool {
	return l.writtenSize > 0 && (time.Now().Hour() == 0 || time.Now().Day() > l.lastLogTime.Day())
}
