package main

import (
	"github.com/go-toast/toast"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type Notify struct {
	notify toast.Notification
}

var instanceNotify *Notify
var onceNotify sync.Once

func GetInstanceNotify() *Notify {
	onceNotify.Do(func() {
		instanceNotify = &Notify{}
		file := CreateIcon()
		instanceNotify.notify = toast.Notification{
			AppID:   "Ruijie自动连接工具",
			Title:   "",
			Message: "",
			Icon:    file,
			Audio:   toast.Default,
		}
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
	// 隐藏模式下不弹通知
	if hideMode == true {
		return
	}
	if len(msgs) < 1 {
		return
	}
	n.notify.Title = msgs[0]
	if len(msgs) > 1 {
		n.notify.Message = msgs[1]
	} else {
		n.notify.Message = ""
	}

	err := n.notify.Push()
	if err != nil {
		log.Error("发送通知失败", err)
		return
	}

	return
}
