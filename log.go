package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// CustomFormatter 自定义的日志格式
type CustomFormatter struct {
	log.Formatter
}

// Format 实现 logrus.Formatter 接口
func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := time.Now().Format(time.RFC3339)
	filename := filepath.Base(entry.Caller.File)
	line := entry.Caller.Line
	msg := entry.Message
	level := entry.Level
	return []byte(fmt.Sprintf("%s [%s] %s:%d %s\n", timestamp, level, filename, line, msg)), nil
}

func InitLog() {
	// 获取程序所在目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个日志文件
	file, err := os.OpenFile(filepath.Join(dir, "Auto-Ruijie-Tool.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// 设置 log 的输出到这个文件
	log.SetOutput(file)

	// 设置自定义的日志格式
	log.SetFormatter(new(CustomFormatter))

	// 设置 log 的级别为 InfoLevel
	log.SetLevel(log.InfoLevel)

	// Enable the line number logger
	log.SetReportCaller(true)
}
