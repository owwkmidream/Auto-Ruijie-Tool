package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func InitLog() {
	// 创建一个日志文件
	file, err := os.OpenFile("Auto-Ruijie-Tool.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// 设置 log 的输出到这个文件
	log.SetOutput(file)

	// 设置 log 的格式为默认的日志格式
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// 设置 log 的级别为 InfoLevel
	log.SetLevel(log.InfoLevel)
}
