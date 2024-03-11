package main

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type Config struct {
	URL           map[string]string `toml:"url"`
	LoginData     map[string]string `toml:"login_data"`
	LogoutData    map[string]string `toml:"logout_data"`
	ManageParams  map[string]string `toml:"manage_params"`
	Headers       map[string]string `toml:"headers"`
	ManageHeaders map[string]string `toml:"manage_headers"`
	Cookie        map[string]string `toml:"cookie"`
	Options       map[string]string `toml:"options"`
	KickParams    map[string]string `toml:"kick_params"`
}

var instance *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{}
		err := instance.Load()
		if err != nil {
			notify.Send("加载配置文件失败", err.Error())
			panic(err)
		}
	})
	return instance
}

func (c *Config) Load() error {
	_, err := toml.DecodeFile("config.toml", &c)
	log.Info("加载配置文件")
	// IO错误则创建一个新的配置文件
	if err != nil {
		if os.IsNotExist(err) {
			// 创建一个新的配置文件
			log.Info("配置文件不存在，创建一个新的配置文件")
			err := CreateConfig()
			if err != nil {
				log.Error("创建配置文件失败", err)
				return err
			}

			// 重新加载配置文件
			_, err = toml.DecodeFile("config.toml", &c)
			if err != nil {
				log.Error("重新加载配置文件失败", err)
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func CreateConfig() error {
	// 创建一个新的配置文件
	file, err := os.Create("config.toml")
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// 使用默认的配置初始化文件
	defaultConfig := GetDefaultConfig()
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(defaultConfig)
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultConfig() Config {
	defaultConfig := Config{
		URL: map[string]string{
			"host":        "http://127.0.0.1",
			"login":       "/eportal/InterFace.do?method=login",
			"logout":      "/eportal/InterFace.do?method=logout",
			"check":       "https://connect.rom.miui.com/generate_204",
			"manage_host": "http://127.0.0.1",
			"get_cookie":  "/selfservice/module/userself/web/portal_business_detail.jsf",
			"get_devices": "/selfservice/module/webcontent/web/onlinedevice_list.jsf",
			"kick":        "/selfservice/module/userself/web/userself_ajax.jsf",
		},
		LoginData: map[string]string{
			"userId":          "",
			"password":        "",
			"service":         "",
			"queryString":     "",
			"operatorPwd":     "",
			"operatorUserId":  "",
			"validcode":       "",
			"passwordEncrypt": "",
		},
		LogoutData: map[string]string{},
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		ManageHeaders: map[string]string{
			"Referer": "http://127.0.0.1",
		},
		Cookie:  map[string]string{},
		Options: map[string]string{},
		ManageParams: map[string]string{
			"channel":  "",
			"name":     "",
			"password": "",
			"ip":       "",
			"callBack": "",
			"index":    "",
		},
	}
	return defaultConfig
}

func (c *Config) Save() error {
	file, err := os.Create("config.toml")
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	encoder := toml.NewEncoder(file)
	return encoder.Encode(c)
}
