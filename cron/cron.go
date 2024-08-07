package cron

import (
	//packexit "go_code/bngo/exit"
	"fmt"
	"go_code/myselfgo/inits"
	"sync"
)

// startBackgroundTasks 启动后台任务
func StartBackgroundTasks() {
	intervalCfg := inits.Config.Interval
	go checkExit(intervalCfg.CheckExit)

}

func StartCronTasks() error {
	intervalCfg := inits.Config.Interval

	num := 3
	wg := &sync.WaitGroup{}
	wg.Add(num)
	ch := make(chan error, num)

	// 交易规范
	go syncPriceFilterInfo(intervalCfg.PriceFilterInfo, wg, ch)
	// 1500s执行一次
	go syncListenKey(intervalCfg.ListenKey, wg, ch)
	//7200秒刷新涨幅榜交易对
	go syncSymbolUp(intervalCfg.SymbolRefresh, wg, ch)
	//市场情绪
	// go syncMarketRate(60, wg, ch)
	//当前持仓1800秒刷新一次
	// go syncRetain(1800, wg, ch)
	// 持仓风险，60秒查询一次
	// go syncPositionRisk(60, wg)
	// 委托单查询
	// go syncWtOrder(60, wg)
	// go syncKlinesMacd(5, wg)
	//go syncHttpKlinesContrastServer(3, wg, ch)
	// 1s执行一次
	//go syncNetworkDelay(intervalCfg.NetworkDelay, wg, ch
	// 判断挂单信息，如果2秒内没有挂单，就取消挂单
	//go SyncCheckOrder(5, wg, ch)

	wg.Wait()
	if len(ch) != 0 {
		return fmt.Errorf(inits.ErrorCronMsg, len(ch))
	}
	close(ch)
	return nil
}
