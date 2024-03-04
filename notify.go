//go:build !windows

package main

import (
	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type Notify struct {
	icon string
}

var instanceNotify *Notify
var onceNotify sync.Once

func GetInstanceNotify() *Notify {
	onceNotify.Do(func() {
		instanceNotify = &Notify{}
		file := CreateIcon()
		instanceNotify.icon = file
	})
	return instanceNotify
}

func CreateIcon() string {
	data, err := Asset("assets/online.png")
	if err != nil {
		log.Error("读取图标文件失败", err)
		return ""
	}

	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "icon-*.png")
	if err != nil {
		log.Error("创建临时文件失败", err)
	}

	if _, err := tmpfile.Write(data); err != nil {
		log.Error("写入临时文件失败", err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Error("关闭临时文件失败", err)
	}

	iconPath := tmpfile.Name()
	return iconPath
}

func (n *Notify) Send(msgs ...string) {
	var title, message string
	if len(msgs) < 1 {
		return
	}
	title = msgs[0]
	if len(msgs) > 1 {
		message = msgs[1]
	}

	err := beeep.Notify(title, message, n.icon)
	if err != nil {
		log.Error("发送通知失败", err)
		return
	}

	return
}
