package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var config *Config
var notify *Notify

// Login 函数用于执行登录操作
func Login() bool {
	data := url.Values{}
	// 遍历配置中的登录数据，将其添加到url.Values对象中
	for key, value := range config.LoginData {
		data.Set(key, value)
	}

	client := &http.Client{}
	// 创建一个新的HTTP请求，方法是POST，URL是配置中的主机地址加上登录路径，请求体是编码后的登录数据
	req, err := http.NewRequest("POST", config.URL["host"]+config.URL["login"], strings.NewReader(data.Encode()))
	// 如果在创建请求时发生错误，返回nil和错误
	if err != nil {
		log.Error("创建请求时发生错误", err)
		return false
	}

	// 遍历配置中的头部数据，将其添加到请求的头部
	for key, value := range config.Headers {
		req.Header.Add(key, value)
	}
	// 使用客户端发送请求，获取响应
	resp, err := client.Do(req)
	if err != nil {
		log.Error("发送请求时发生错误", err)
		return false
	}

	// 读取响应体
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

	return res == "success"
}

//func Logout() (*http.Response, error) {
//	client := &http.Client{}
//	req, err := http.NewRequest("POST", config.URL["host"]+config.URL["logout"], nil)
//	if err != nil {
//		return nil, err
//	}
//
//	for key, value := range config.Headers {
//		req.Header.Add(key, value)
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//
//	return resp, nil
//}

func TestNet(url string) bool {
	resp, err := http.Get(url)
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}
	if err != nil {
		log.Error("检测校园网环境出错", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("关闭Body出错", err)
		}
	}(resp.Body)

	return false
}

func main() {
	InitLog()
	log.Info("程序启动--------------------------------------")
	notify = GetInstanceNotify()
	config = GetInstance()

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
