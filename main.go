package main

func main() {
	// 初始化控制台
	InitFlag()
	// 开始安全退出任务
	InitSafeExit()
	// 初始化配置
	InitConf(configPath)
	// 初始化日志
	InitLog()
	// 初始化断点
	InitBreakPoint()
	// 开始任务
	InitTask()
}
