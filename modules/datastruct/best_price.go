package datastruct

import (
	"sync"
	"time"

	"go_code/myselfgo/define"
	"go_code/myselfgo/utils"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/open-binance/logger"
)

type SafeBestPriceMap struct {
	L *sync.RWMutex            `json:"-"`
	M map[string]BestPriceInfo `json:"m"` //key：交易对
}

// BestPriceInfo 最优惠价格信息
type BestPriceInfo struct {
	Updatetime   int64   `json:"ut"`   // 这个字段存储了最后一次更新的时间，单位是毫秒
	UpdateID     int64   `json:"u_id"` // 这个字段存储了从Binance获取的更新ID
	Symbol       string  `json:"s"`    // 交易对
	BestBidPrice float64 `json:"bbp"`  // 最佳买入价格
	BestBidQty   float64 `json:"bbq"`  // 最佳买入数量
	BestBidValue float64 `json:"bbv"`  // 最佳买入价值的总额
	BestAskPrice float64 `json:"bap"`  // 最佳卖出价格
	BestAskQty   float64 `json:"baq"`  // 最佳卖出数量
	BestAskValue float64 `json:"bav"`  // 最佳卖出价值的总额
}

// ReInit reinits the map, just for test
func (sm *SafeBestPriceMap) ReInit(m map[string]BestPriceInfo) {
	sm.L.Lock()
	defer sm.L.Unlock()
	sm.M = m
}

/*
 * @description: 更新最优价格信息
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:06:51
 * @returns: true  说明成功
 */
func (sm *SafeBestPriceMap) Update(event *futures.WsBookTickerEvent) bool {
	if event == nil {
		return false
	}

	// 获取更新ID和交易对标识
	updateID := event.UpdateID
	symbol := event.Symbol

	// 获取当前时间戳
	nowTs := time.Now().UnixMilli()

	// 将字符串转换为浮点数
	bbp, err := utils.StrTofloat64(event.BestBidPrice)
	if err != nil {
		logger.Errorf("无效的最优买价，bbp=%s，err=%s", event.BestBidPrice, err.Error())
		return false
	}

	bbq, err := utils.StrTofloat64(event.BestBidQty)
	if err != nil {
		logger.Errorf("无效的最优购买数量，bbq=%s，err=%s", event.BestBidQty, err.Error())
		return false
	}

	bap, err := utils.StrTofloat64(event.BestAskPrice)
	if err != nil {
		logger.Errorf("无效的最优卖价，bap=%s，err=%s", event.BestAskPrice, err.Error())
		return false
	}

	baq, err := utils.StrTofloat64(event.BestAskQty)
	if err != nil {
		logger.Errorf("无效的最优卖数量，baq=%s，err=%s", event.BestAskQty, err.Error())
		return false
	}

	// 加锁保证线程安全
	sm.L.Lock()
	defer sm.L.Unlock()

	// 创建BestPriceInfo结构体实例，并设置其属性值
	bestPrice := BestPriceInfo{
		Updatetime:   nowTs,
		UpdateID:     updateID,
		Symbol:       symbol,
		BestBidPrice: bbp,
		BestBidQty:   bbq,
		BestAskPrice: bap,
		BestAskQty:   baq,
	}

	// 将BestPriceInfo实例保存到map中
	sm.M[symbol] = bestPrice

	return true
}

/*
 * @description: 根据交易对获取最优的价格信
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:10:31
 */
func (sm *SafeBestPriceMap) GetBySymbol(symbol string) (BestPriceInfo, bool) {
	sm.L.RLock()
	defer sm.L.RUnlock()
	bestPrice, ok := sm.M[symbol]
	if !ok {
		return bestPrice, false
	}

	return bestPrice, true
}

/*
 * @description: 得到所有的symbol数量
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:11:13
 */
func (sm *SafeBestPriceMap) GetSymbolCount() int {
	sm.L.RLock()
	defer sm.L.RUnlock()
	return len(sm.M)
}

/*
 * @description: 必须深度复制地图以避免数据竞争
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:13:17
 */
func (sm *SafeBestPriceMap) Get() map[string]BestPriceInfo {
	sm.L.RLock()
	defer sm.L.RUnlock()
	m := make(map[string]BestPriceInfo)
	for symbol, info := range sm.M {
		m[symbol] = info
	}
	return m
}

/*
 * @description: 删除某个交易对的信息，必须深度复制地图以避免数据竞争
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:14:26
 */
func (sm *SafeBestPriceMap) DeleteBySymbol(key string) {
	sm.L.Lock()
	defer sm.L.Unlock()
	delete(sm.M, key)
}

/*
 * @description: 将数量转换为价值，如果找不到，返回0值
 * @fileName: best_price.go
 * @author: vip120@126.com
 * @date: 2024-03-25 15:15:36
 */
func ConvertQty2Value(m map[string]BestPriceInfo, base string, qty float64) float64 {
	// 定义 USDT 的别名
	usdt := define.QuoteUSDT

	// 如果基础货币等于 USDT，则直接返回数量
	if base == usdt {
		return qty
	}

	// 拼接基础货币和 USDT 的符号
	symbol := utils.JoinStrWithSep("", base, usdt)

	// 从 map 中获取对应符号的最佳价格信息
	bestPrice, ok := m[symbol]

	// 如果 map 中不存在该符号的最佳价格信息，则返回 0
	if !ok {
		return 0
	}

	// 返回数量乘以最佳买入价格
	return qty * bestPrice.BestBidPrice
}
