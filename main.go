package main

import (
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

		// 包含"电脑"or"路由器"，则踢出
		if strings.Contains(label, "电脑") || strings.Contains(label, "路由器") {
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
	config = &Config{
		URL: map[string]string{
			"host":        "http://210.27.177.172",
			"manage_host": "http://210.27.185.13:8080",
			"get_cookie":  "/selfservice/module/userself/web/portal_business_detail.jsf",
			"get_devices": "/selfservice/module/webcontent/web/onlinedevice_list.jsf",
			"kick":        "/selfservice/module/userself/web/userself_ajax.jsf",
		},
		LoginData: map[string]string{
			"userId": "210809010107",
		},
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		ManageHeaders: map[string]string{
			"Referer": "http://210.27.177.172/",
		},
		ManageParams: map[string]string{
			"channel":  "cG9ydGFs",
			"name":     "d281ac3212ae6c72b261b19698561587",
			"password": "7a788f8f412886019a1acacb8f95e9ad",
			"ip":       "210.27.177.172",
			"callBack": "portal_business_detail",
			"index":    "3",
		},
		KickParams: map[string]string{
			"methodName": "indexBean.kickUserBySelfForAjax",
		},
		Options: map[string]string{
			"timeout": "60",
		},
	}

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
			time.Sleep(1 * time.Second)
		}
		if !res {
			log.Fatal("当前不处于校园网环境，退出程序")
			notify.Send("校园网环境异常", "当前不处于校园网环境，退出程序")
			return
		}
	}
	log.Info("当前处于校园网环境")

	log.Info("当前为踢出模式")
	// 断网
	KickDevices()
}
