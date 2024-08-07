package datastruct

import (
	"fmt"
	"sync"
	"time"

	"go_code/myselfgo/define"
	"go_code/myselfgo/utils"

	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

type SafeCumulativeReturnRate struct {
	L                    *sync.RWMutex   `json:"-"`
	CumulativeQuoteValue decimal.Decimal `json:"cumulative_quote_value"`
	TotalDeltaValue      decimal.Decimal `json:"total_delta_value"`
	TotalReturnRate      decimal.Decimal `json:"total_return_rate"` // assigned in CalcReturnRate function
}

// CalcReturnRate 根据预期价格、实际价格、预期收益率、价值和累计值计算实际收益率和总收益率，并更新累计报价价值、总价值变动量和总收益率
// expectedPrice: 预期价格
// realPrice: 实际价格
// expectedReturnRate: 预期收益率
// value: 价值
// cumulative: 累计值
// traceID: 跟踪ID
func (sr *SafeCumulativeReturnRate) CalcReturnRate(expectedPrice, realPrice, expectedReturnRate,
	value, cumulative decimal.Decimal, traceID string) {
	// 计算价格变动率
	deltaPriceRate := realPrice.Sub(expectedPrice).Div(expectedPrice)

	// 计算实际收益率
	realReturnRate := expectedReturnRate.Sub(define.Decimal1).Add(deltaPriceRate)

	// 计算价值变动量
	deltaValue := realReturnRate.Mul(value)

	// 加锁
	sr.L.Lock()
	// 解锁
	defer sr.L.Unlock()

	// 计算总价值变动量
	totalDeltaValue := sr.TotalDeltaValue.Add(deltaValue)

	// 计算总收益率
	totalReturnRate := totalDeltaValue.Div(cumulative).Mul(define.Decimal10000)

	// 更新累计报价价值
	sr.CumulativeQuoteValue = cumulative
	// 更新总价值变动量
	sr.TotalDeltaValue = totalDeltaValue
	// 更新总收益率
	sr.TotalReturnRate = totalReturnRate

	// 记录日志
	logger.Infof("msg=%s||real_return_rate=%s||total_delta_value=%v||cumulative=%v||total_return_rate=%v||trace_id=%s",
		"succeed to calc total return rate", realReturnRate, totalDeltaValue, cumulative, totalReturnRate, traceID)
}

// GetReturnRate 返回总收益率和总收益值
//
// 参数：
//
//	sr *SafeCumulativeReturnRate - SafeCumulativeReturnRate结构体的指针
//
// 返回值：
//
//	decimal.Decimal - 总收益率
//	decimal.Decimal - 总收益值
func (sr *SafeCumulativeReturnRate) GetReturnRate() (decimal.Decimal, decimal.Decimal) {
	// 加锁，确保并发安全
	sr.L.Lock()
	// 函数执行完毕后解锁
	defer sr.L.Unlock()

	// 返回总收益率和总收益值
	return sr.TotalReturnRate, sr.TotalDeltaValue
}

func (sr *SafeCumulativeReturnRate) GetAll() SafeCumulativeReturnRate {
	// 加锁，保证并发安全
	sr.L.Lock()
	// 函数返回前解锁
	defer sr.L.Unlock()

	// 返回一个新的 SafeCumulativeReturnRate 实例，包含当前实例的字段值
	return SafeCumulativeReturnRate{
		CumulativeQuoteValue: sr.CumulativeQuoteValue,
		TotalDeltaValue:      sr.TotalDeltaValue,
		TotalReturnRate:      sr.TotalReturnRate,
	}
}

type SafeDecimal struct {
	L          *sync.RWMutex
	Value      float64
	UpdateTime int64 // unit: ms
}

// Update 更新SafeDecimal的值，返回是否更新成功
func (sd *SafeDecimal) Update(value float64) bool {
	// 加锁，确保并发安全
	sd.L.Lock()
	defer sd.L.Unlock()

	// 获取当前时间毫秒数
	updateTime := time.Now().UnixMilli()
	// 获取上次更新时间
	lastUpdateTime := sd.UpdateTime
	// 计算时间差
	delta := updateTime - lastUpdateTime
	// 判断是否需要更新
	update := delta <= 2222

	// 输出日志，显示当前时间、上次更新时间、时间差和是否需要更新
	logger.Infof("msg=%s||current=%d||last=%d||delta=%d||update=%t",
		"display time in Update function", updateTime, lastUpdateTime, delta, update)

	// 如果不需要更新，则直接返回false
	if update {
		logger.Infof("msg=%s||current=%d||last=%d||delta=%d||update=%t",
			"do nothing in Update function", updateTime, lastUpdateTime, delta, update)
		return false
	}

	// 更新Value和UpdateTime字段
	sd.Value = value
	sd.UpdateTime = updateTime

	// 返回true，表示更新成功
	return true
}

// Set sets the value
func (sd *SafeDecimal) Set(value float64) {
	sd.L.Lock()
	defer sd.L.Unlock()

	sd.Value = value
	sd.UpdateTime = time.Now().UnixMilli()
}

// Get gets the value
func (sd *SafeDecimal) Get() float64 {
	sd.L.RLock()
	defer sd.L.RUnlock()

	return sd.Value
}

// SafeAssetInfo specifies a safe map for asset info
type SafeAssetInfoMap struct {
	L *sync.RWMutex
	M map[string]AssetInfo // key: base asset, value: asset info
}

// AssetInfo specifies the account info for a asset
type AssetInfo struct {
	Asset      string  `json:"asset,omitempty"`
	UpdateTime int64   `json:"update_time,omitempty"` // account update time, unit: ms
	Free       float64 `json:"free,omitempty"`
	FreeValue  float64 `json:"free_value,omitempty"`
	Locked     float64 `json:"locked,omitempty"`
}

// ReInit reinits the map
func (sm *SafeAssetInfoMap) ReInit(m map[string]AssetInfo, exitThrd float64) bool {
	// 初始化退出标志为false
	exit := false
	// 遍历传入的map
	for asset, assetInfo := range m {
		// 如果资产为BNB
		if asset == define.AssetBNB {
			// 获取BNB的可用数量
			freeDecimal := assetInfo.Free
			// 如果可用数量小于退出阈值
			if freeDecimal < exitThrd {
				// 打印日志，提示因为BNB的可用数量不足而退出
				logger.Infof("msg=%s||asset=%s||free=%v||exit_thrd=%v",
					"the progress will exit because free of BNB", asset, freeDecimal, exitThrd)
				// 设置退出标志为true
				exit = true
			}
		}
	}

	// 加锁
	sm.L.Lock()
	defer sm.L.Unlock()
	// 更新SafeAssetInfoMap的M字段为传入的m
	sm.M = m
	// 返回退出标志
	return exit
}

// UpdateFreeValue 更新指定资产的FreeValue值
// 参数：
//   - asset: 待更新FreeValue值的资产名称
//   - freeValue: 待更新的FreeValue值
//
// 返回值：
//   - error: 更新成功返回nil，否则返回错误信息
func (sm *SafeAssetInfoMap) UpdateFreeValue(asset string, freeValue float64) error {
	// 加锁，确保并发安全
	sm.L.Lock()
	defer sm.L.Unlock()

	// 从映射中获取指定资产的资产信息
	assetInfo, ok := sm.M[asset]
	if !ok {
		// 如果获取失败，则返回错误信息
		return fmt.Errorf("get no asset info with '%s'", asset)
	}

	// 更新资产信息的 FreeValue 属性
	assetInfo.FreeValue = freeValue
	// 将更新后的资产信息重新存入映射中
	sm.M[asset] = assetInfo
	// 返回空，表示更新成功
	return nil
}

// Update updates the asset info
// return (buy bnb or not, succeed or not)
/*
 * @description: 更新账户信息
 * @fileName: asset.go
 * @author: vip120@126.com
 * @date: 2024-04-16 10:22:27
 */
// func (sm *SafeAssetInfoMap) Update(accountUpdateTime int64, au futures.WsBalance) (bool, bool) {

// 	// 是否需要退出
// 	exit := false
// 	// 资产
// 	asset := au.Asset
// 	// 钱包余额
// 	balance, err := decimal.NewFromString(au.Balance)
// 	if err != nil {
// 		logger.Errorf("msg=invalid balance||balance=%s||err=%s", balance, err.Error())
// 		return exit, false
// 	}
// 	// 锁定数量
// 	brossWalletBalance, _ := decimal.NewFromString(au.CrossWalletBalance) //除去逐仓仓位保证金的钱包余额
// 	changeBalance, _ := decimal.NewFromString(au.ChangeBalance)           //除去盈亏与交易手续费以外的钱包余额改变量
// 	// 加锁
// 	sm.L.Lock()
// 	defer sm.L.Unlock()

// 	// 创建新的资产信息
// 	assetInfo := AssetInfo{
// 		Asset:      asset,
// 		UpdateTime: accountUpdateTime,
// 		Free:       freeDecimal,
// 		FreeValue:  freeValue,
// 		Locked:     lockedDecimal,
// 	}

// 	// 更新资产信息
// 	sm.M[asset] = assetInfo

// 	// 返回结果
// 	return exit, true
// }

// GetByAsset 从SafeAssetInfoMap中根据资产名称获取AssetInfo
// 参数：
//
//	sm *SafeAssetInfoMap：SafeAssetInfoMap对象指针
//	asset string：需要查询的资产名称
//
// 返回值：
//
//	AssetInfo：查询到的资产信息
//	bool：是否成功查询到资产信息
func (sm *SafeAssetInfoMap) GetByAsset(asset string) (AssetInfo, bool) {
	// 读取共享锁
	sm.L.RLock()
	// 在函数返回前释放共享锁
	defer sm.L.RUnlock()

	// 从map中获取指定asset的AssetInfo
	assetInfo, ok := sm.M[asset]
	// 如果map中不存在该asset的AssetInfo
	if !ok {
		// 返回空的AssetInfo和false
		return assetInfo, false
	}

	// 返回找到的AssetInfo和true
	return assetInfo, true
}

// Get gets all the asset info
func (sm *SafeAssetInfoMap) Get() map[string]AssetInfo {
	sm.L.RLock()
	defer sm.L.RUnlock()

	m := make(map[string]AssetInfo)
	for symbol, info := range sm.M {
		m[symbol] = info
	}
	return m
}

// GetMin 从SafeAssetInfoMap中获取可用价值最小的资产和其资产信息
// 返回值为最小资产的字符串表示和对应的AssetInfo
func (sm *SafeAssetInfoMap) GetMin() (string, AssetInfo) {
	sm.L.RLock()
	defer sm.L.RUnlock()

	// 将USDT设置为最小资产
	minAsset := define.CoinUSDT
	min := sm.M[minAsset]

	// 遍历资产映射表
	for asset, assetInfo := range sm.M {
		// 如果资产不在HoldCoinMap中，则跳过该资产
		if _, ok := define.HoldCoinMap[asset]; !ok {
			continue
		}

		// 如果当前资产的可用价值小于min的可用价值，则更新最小资产和对应的资产信息
		if assetInfo.FreeValue < min.FreeValue {
			minAsset = asset
			min = assetInfo
		}
	}

	// 返回最小资产和对应的资产信息
	return minAsset, min
}

// Qty2value 根据资产、数量和最佳价格映射计算资产价值
//
// 参数：
// asset: 资产符号
// qty: 数量
// bpm: 最佳价格映射
//
// 返回值：
// float64: 资产价值
// error: 错误信息
func Qty2value(asset string, qty float64, bpm *SafeBestPriceMap) (float64, error) {
	// 判断资产是否为稳定币
	isStable := utils.IsStableCoin(asset)

	// 如果是稳定币，则直接返回数量
	if isStable {
		return qty, nil
	}

	var value float64
	// 拼接资产和USDT的符号
	symbol := utils.JoinStrWithSep("", asset, define.QuoteUSDT)
	// 从最佳价格映射中获取对应符号的最佳价格
	bestPrice, ok := bpm.GetBySymbol(symbol)
	if !ok {
		// 获取失败，则返回错误
		return value, fmt.Errorf("get no best price, asset: %s, symbol: %s", asset, symbol)
	}
	// 计算价值
	value = bestPrice.BestBidPrice * qty

	// 成功返回价值
	// succeed
	return value, nil
}

func qty2value(special bool, qty float64, bp BestPriceInfo) float64 {
	// 判断是否为特殊商品
	if special {
		// 如果是特殊商品，则使用最佳卖出价计算价值
		return qty / bp.BestAskPrice
	}
	// 如果不是特殊商品，则使用最佳买入价计算价值
	return qty * bp.BestBidPrice
}
