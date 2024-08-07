package main

import (
	"flag"
	"fmt"
	"go_code/myselfgo/client"
	"go_code/myselfgo/cron"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/notice"
	"go_code/myselfgo/utils/spot"
	"go_code/myselfgo/wsclient"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-binance/logger"
)

/*
 * @description: 初始化函数
 * @fileName: main.go
 * @author: vip120@126.com
 * @date: 2024-04-10 17:57:04
 */

func init() {
	// 初始化配置
	initConfig()
	// 初始化日志记录器
	initLogger()
	// 初始化代理
	initProxy()
	// 获取服务器时间
	getServerTime()
	// 初始化HTTP连接池
	initHttpPool()

}

func main() {
	// 调用startProcess函数
	startProcess()
	// cron.GetMarketRate()
	//清仓
	// cron.ClearOrderServer()
	//持仓
	// cron.ShowOrderServer()
	//测试
	// cron.Test()
}

// initConfig 加载配置文件
func initConfig() {
	// 定义一个字符串指针变量cfgFile，用于存储配置文件的路径
	cfgFile := flag.String("c", "configs/cfg.local.yaml", "配置文件")
	// 解析命令行参数
	flag.Parse()

	// 调用inits包的LoadConfig函数加载配置文件，传入cfgFile指针指向的值作为参数
	if err := inits.LoadConfig(*cfgFile); err != nil {
		// 如果加载配置文件出现错误，调用inits包的ErrorMsg函数输出错误信息
		inits.ErrorMsg(inits.LoadConfigError, err)
	}
	// 调用inits包的SuccessInfoMsg函数输出成功加载配置文件的提示信息
	inits.SuccessInfoMsg(inits.SuccessLoadConfig)
}

// initLogger 初始化日志对象
func initLogger() {
	// 创建一个默认的日志配置对象
	logCfg := logger.NewDefaultCfg()

	// 使用默认配置初始化日志
	if err := logger.Init(logCfg); err != nil {
		// 如果初始化失败，输出错误信息
		inits.ErrorMsg(inits.InitLoggerError, err)
	}

	// 在函数结束时，确保所有待写入的日志都已写入
	defer logger.Sync()

	// 使用文件配置初始化文件日志
	if err := inits.InitFileLogger(inits.Config.File); err != nil {
		// 如果文件日志初始化失败，输出错误信息
		inits.ErrorMsg(inits.InitLoggerFileError, err)
	}

	// 初始化日志成功，输出成功信息
	inits.SuccessInfoMsg(inits.SuccessInitLoger)
}

// initProxy 设置网络代理
func initProxy() {
	if err := inits.DoSetProxy(); err != nil {
		inits.ErrorMsg(inits.DoSetProxyError, err)
	}
	inits.SuccessInfoMsg(inits.SuccessProxy)
}

// getServerTime 获取服务器时间
func getServerTime() {
	account := inits.Config.Account
	timestamp, err := client.GetServerTime(account.ApiKey, account.SecretKey)
	if err != nil {
		logger.Errorf(inits.MsgGetServerTime, inits.ErrorGetServerTime, account.ApiKey, err.Error())
		inits.ErrorMsg(inits.ErrorGetServerTime, err)
		client.GracefulExit(err.Error())
	}
	logger.Infof(inits.MsgGetServerTime, inits.SuccessGetServerTime, account.ApiKey, timestamp)
	inits.SuccessInfoMsg(inits.SuccessGetServerTime)
}

// initHttpPool 初始化client链接池
func initHttpPool() {
	if err := client.Init(inits.Config); err != nil {
		inits.ErrorMsg(inits.ErrorInitHttpPool, err)
	}
	inits.SuccessInfoMsg(inits.SuccessInitHttpPool)
}

/*
 * @description: 主进程函数
 * @fileName: main.go
 * @author: vip120@126.com
 * @date: 2024-03-23 09:41:24
 */
// startProcess 是主函数，负责启动整个流程
func startProcess() {
	go cron.StartCronTasks()
	time.Sleep(3 * time.Second)
	go cron.StartHttpKlinesRootSymbol()
	if inits.Config.Symbol.Type == 2 {
		go cron.StartHttpKlinesServe2()
	} else {
		fmt.Println("请在配置文件中填写正确的交易类型，1表示现货交易，2表示合约交易")
	}

	go wsclient.StartWSProcess()
	var startMsg string
	if inits.Config.Symbol.Type == 1 {
		startMsg = "现货反包策略预警策略--启动，请等待通知。"
	} else {
		// 判断策略类型
		var side, w string
		if inits.Config.Order.Side == 1 {
			side = "只做多"
		} else if inits.Config.Order.Side == 2 {
			side = "只做空"
		} else if inits.Config.Order.Side == 3 {
			side = "做多做空"
		}
		if inits.Config.Order.Warehouse_mode == 1 {
			w = "仓位模式：单仓"
		} else if inits.Config.Order.Warehouse_mode == 2 {
			w = "仓位模式：多仓"
		}
		startMsg = "合约反包策略预警策略--启动，请等待通知。\n策略类型：" + side + "\n" + w
	}
	notice.SendDingTalk(startMsg)

	waitForExitSignal()
}

// waitForExitSignal 等待退出信号
func waitForExitSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	fmt.Println("已退出信号:", s)
	logger.Infof("已退出信号,signal=%+v", s)

	var startMsg string
	if inits.Config.Symbol.Type == 1 {
		startMsg = "现货反包策略预警策略--退出"
	} else {
		startMsg = "合约反包策略预警策略--退出"
	}
	notice.SendDingTalk(startMsg)
	spot.GracefulExit(fmt.Sprintf("已退出信号: %+v", s))
}
