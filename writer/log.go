package writer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	MaxLogFileSizeInBytes = 5000000
)

type LogWriter interface {
	Err(string)
	Flush() error
}

type FileLogWriter struct {
	logFileName string
	logs        []string
}

func NewLogWriter() (LogWriter, error) {
	l := &FileLogWriter{}
	dirName, err := getEffingoFolderPath()
	if err != nil {
		return nil, err
	}

	l.logFileName = filepath.Join(dirName, ".effingo_log")
	l.logs = make([]string, 0)

	return l, nil
}

// Err receive a formatted string and keep in the logs
// to be later flushed
func (l *FileLogWriter) Err(err string) {
	l.logs = append(l.logs, err)
}

// Flush will write to the logFileName
// the contents inside logs
func (l *FileLogWriter) Flush() error {
	file, err := os.OpenFile(l.logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Printf("Error opening log file %v\n", err)
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := info.Size()
	if fileSize > MaxLogFileSizeInBytes {
		if err := file.Truncate(0); err != nil {
			return err
		}
	}

	var logToBeWritten string
	for _, log := range l.logs {
		logToBeWritten += l.formatLogMessage(log)
	}

	if _, err := file.WriteString(logToBeWritten); err != nil {
		log.Printf("Error writing to string %v\n", err)
		return err
	}

	return nil
}

// formatLogMessage format the error message received
func (l *FileLogWriter) formatLogMessage(msg string) string {
	return fmt.Sprintf("%v - %v", time.Now().String(), msg)
}
