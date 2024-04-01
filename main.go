package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/gofrs/flock"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func KickDevices() bool {
	// 首先获取cookies
	client := &http.Client{
		Transport: http.DefaultTransport,
	}

	req, err := http.NewRequest("GET", config.URL["manage_host"]+config.URL["get_cookie"], nil)
	if err != nil {
		log.Error("创建请求时发生错误", err)
		return false
	}
	// 添加ManageHeaders到请求头
	for key, value := range config.ManageHeaders {
		req.Header.Add(key, value)
	}

	// 把ManageParams添加到get请求的URL中
	query := req.URL.Query()
	for key, value := range config.ManageParams {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Error("发送请求时发生错误", err)
		return false
	}

	// 获取cookies
	cookies := resp.Cookies()
	defer resp.Body.Close()

	// 使用cookies获取设备列表
	req, err = http.NewRequest("GET", config.URL["manage_host"]+config.URL["get_devices"], nil)
	if err != nil {
		log.Error("创建请求时发生错误", err)
		return false
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	//发送请求，解析响应的html
	resp, err = client.Do(req)
	if err != nil {
		log.Error("发送请求时发生错误", err)
		return false
	}
	bodyReader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		log.Error("读取响应体时发生错误", err)
		return false
	}
	defer resp.Body.Close()

	bodyString := string(bodyBytes)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	if err != nil {
		log.Error("解析html时发生错误", err)
		return false
	}

	// 获取#mydiv > div[id*="divId"]
	divs := doc.Find("div[id*=\"divId\"]")
	divs.Each(func(i int, selection *goquery.Selection) {
		// 得到div的nodeList，获取子元素
		// #a1得到IP地址
		ip := selection.Find("#a1").Text()
		ip = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`).FindString(ip)

		// label得到设备名称
		label := selection.Find("label").Text()
		label = strings.TrimSpace(label)

		log.Info("检测到设备：", label, "ip: ", ip)

		// label不包含"电脑"or"路由器"，则踢出
		if !strings.Contains(label, "电脑") && !strings.Contains(label, "路由器") && !strings.Contains(label, "校园网") {
			log.Info("踢出设备：", label, " ip: ", ip)
			// 踢出设备
			req, err = http.NewRequest("POST", config.URL["manage_host"]+config.URL["kick"], nil)
			// 添加kick params
			query = req.URL.Query()
			for key, value := range config.KickParams {
				query.Add(key, value)
			}
			req.URL.RawQuery = query.Encode()
			// 添加body
			body := strings.NewReader(url.Values{
				"key": {config.LoginData["userId"] + ":" + ip},
			}.Encode())
			req.Body = io.NopCloser(body)
			req.ContentLength = int64(body.Len())

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}

			// 发送请求
			resp, err = client.Do(req)
			if err != nil {
				log.Error("发送请求时发生错误", err)
				return
			}
			defer resp.Body.Close()

			log.Infof("踢出设备 %v 成功", label)
		}
	})

	return true
}

func Login() bool {
	for i := 0; i < 3; i++ {
		KickDevices()
		if SendRequest("login", config.LoginData) {
			return true
		}
	}
	return false
}

func Logout() bool {
	return SendRequest("logout", config.LogoutData)
}

func TestNet(url string) bool {
	client := &http.Client{
		Timeout: time.Second * 1,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Errorf("检测环境出错：%v %v", url, err)
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
	timeout, _ := strconv.Atoi(config.Options["timeout"])

	if !res {
		log.Info("校园网环境异常，循环检测", timeout, "秒")
		notify.Send("校园网环境异常，循环检测", strconv.Itoa(timeout), "秒")
		for i := 1; i <= timeout; i++ {
			log.Info("进行第", i, "次检测")
			if res = TestNet(config.URL["host"]); res {
				break
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

		return
	}

	// 登录
	if res = Login(); res {
		log.Info("登录成功")
		notify.Send("登录成功")
	} else {
		log.Error("登录失败")
		notify.Send("登录失败", "请检查日志")
	}
}
