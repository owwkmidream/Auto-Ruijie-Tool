# Auto-Ruijie-Tool
# 自动锐捷认证工具

## 特点
工具基于GO语言开发，支持Windows、Linux、MacOS操作系统


## TODO
- [x] 支持自动登录
- [x] 支持日志
- [x] 支持自动重试
- [x] 支持其他平台
- [x] 支持自动断网
- [x] 支持自动踢出其他设备
- [ ] 支持配置是否自动断网、自动踢出其他设备（大概率不会支持）
- [ ] 支持命令行独立运行功能（大概率不会支持）

## 使用
运行一次后会自动生成config.toml，请抓包之后填写相关参数
一般填写URL.host和LoginData即可

使用自动踢出需要抓包对应路径的params或者body，具体看config.toml里的默认URL

## 编译
go build -o Auto-Ruijie-Tool.exe Auto-Ruijie-Tool