# Auto-Ruijie-Tool
# 自动锐捷认证工具

## 特点
工具基于GO语言开发，支持Windows、Linux、MacOS操作系统


## TODO
- [x] 支持自动登录
- [x] 支持日志
- [x] 支持自动重试
- [x] 支持其他平台
- [ ] 支持自动断网
- [ ] 支持自动踢出其他设备

## 使用
运行一次后会自动生成config.toml，请抓包之后填写相关参数
一般填写URL.host和LoginData即可

## 编译
go build -o Auto-Ruijie-Tool.exe Auto-Ruijie-Tool