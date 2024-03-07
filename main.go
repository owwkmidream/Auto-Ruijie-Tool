package main

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/flock"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var config *Config
var notify *Notify

// SendRequest 函数用于执行请求
func SendRequest(endpoint string, requestData map[string]string) bool {
	data := url.Values{}
	for key, value := range requestData {
		data.Set(key, value)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", config.URL["host"]+config.URL[endpoint], strings.NewReader(data.Encode()))
	if err != nil {
		log.Error("创建请求时发生错误", err)
		return false
	}

	for key, value := range config.Headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("发送请求时发生错误", err)
		return false
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取响应体时发生错误", err)
		return false
	}
	defer resp.Body.Close()

	// 解析响应体为一个 map
	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Error("解析响应体时发生错误", err)
		return false
	}

	// 获取 result 字段
	res, _ := result["result"]
	log.Info("响应：", result)
	if res == "success" {

		return true
	} else {
		log.Error("请求体：", resp.Request)
		return false
	}
}

func Login() bool {
	return SendRequest("login", config.LoginData)
}

func Logout() bool {
	return SendRequest("logout", config.LogoutData)
}

func TestNet(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Error("检测环境出错：%v", url, err)
		return false
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("无法关闭body", err)
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}

	return false
}

func tryLockFile(path string) (*flock.Flock, error) {
	fileLock := flock.New(path)
	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("文件已被另一个进程锁定")
	}
	return fileLock, nil
}

func main() {
	InitLog()
	notify = GetInstanceNotify()
	config = GetInstance()

	fileLock, err := tryLockFile("./.ruijie.lock")
	if err != nil {
		notify.Send("运行失败", err.Error())
		log.Fatalf("无法锁定文件: %v", err)
	}
	defer fileLock.Unlock()

	log.Info("程序启动--------------------------------------")

	// 检测校园网环境，状态码200-299表示正常
	log.Info("检测校园网环境")
	res := TestNet(config.URL["host"])
	if !res {
		log.Info("校园网环境异常，循环检测60秒")
		notify.Send("校园网环境异常", "循环检测60秒")
		for i := 1; i <= 60; i++ {
			res = TestNet(config.URL["host"])
			if res {
				break
			} else {
				log.Info("进行第", i, "次检测")
			}
		}
		if !res {
			log.Fatal("当前不处于校园网环境，退出程序")
			notify.Send("校园网环境异常", "当前不处于校园网环境，退出程序")
			return
		}
	}
	log.Info("当前处于校园网环境")

	// 检测是否已经登录
	if TestNet(config.URL["check"]) {
		log.Info("已经登录")
		notify.Send("已经登录", "网络已连接")

		// TODO: 是否需要退出登录
		return
	}

	// 登录
	if res = Login(); res {
		log.Info("登录成功")
		if TestNet(config.URL["check"]) {
			notify.Send("登录成功", "网络已连接")
		} else {
			notify.Send("登录成功", "网络未连接")
		}
	} else {
		log.Error("登录失败")
		notify.Send("登录失败", "请检查日志")
	}
}
