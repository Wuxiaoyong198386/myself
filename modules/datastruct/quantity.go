package datastruct

import (
	"errors"
	"fmt"
	"sync"

	"go_code/myselfgo/define"
	"go_code/myselfgo/utils"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
)

// SafeQuoteQtyInfo 指定报价数量信息
type SafeQuoteQtyInfo struct {
	L                    *sync.RWMutex      `json:"-"`
	PlanQuoteValue       float64            `json:"plan_quote_value"`
	CumulativeQuoteValue float64            `json:"cumulative_quote_value"`
	DeltaQuoteValue      float64            `json:"delta_quote_value"` // PlanQuoteqty - CumulativeQuoteqty
	Commission           map[string]float64 `json:"commission"`        // key: asset, value: commission value
}

// SetValue 设置SafeQuoteQtyInfo实例中的PlanQuoteValue字段，同时重置CumulativeQuoteValue字段，并计算DeltaQuoteValue字段的值
// 参数：
//
//	planValue - 要设置的PlanQuoteValue值
//
// 返回值：
//
//	无
func (si *SafeQuoteQtyInfo) SetValue(planValue float64) {
	si.L.Lock()
	defer si.L.Unlock()
	si.PlanQuoteValue = planValue
	si.CumulativeQuoteValue = 0
	si.DeltaQuoteValue = si.PlanQuoteValue - si.CumulativeQuoteValue
}

// UpdateValueWithResp 根据 Binance 创建订单响应和最佳价格映射更新 SafeQuoteQtyInfo 实例中的值
//
// 参数：
//
//	si *SafeQuoteQtyInfo：SafeQuoteQtyInfo 实例指针
//	resp *binance.CreateOrderResponse：Binance 创建订单响应指针
//	bpm *SafeBestPriceMap：最佳价格映射指针
//
// 返回值：
//
//	bool：是否完成交易（计划报价价值小于等于累计报价价值）
//	error：更新过程中遇到的错误，如果没有错误则为 nil
func (si *SafeQuoteQtyInfo) UpdateValueWithResp(resp *binance.CreateOrderResponse, bpm *SafeBestPriceMap) (bool, error) {
	if resp == nil {
		return false, errors.New("nil response from binance")
	}

	status := resp.Status
	if status != binance.OrderStatusTypeFilled {
		return false, fmt.Errorf("unexpected status from binance, status: %s", status)
	}

	charged := false
	for _, fill := range resp.Fills {
		if fill == nil {
			continue
		}
		commission, err := utils.NewFromString(fill.Commission)
		if err != nil {
			continue
		}
		charged = charged || commission.Cmp(decimal.Zero) > 0
		if charged {
			break
		}
	}
	if !charged {
		return false, fmt.Errorf("trade is free")
	}

	// 将累计报价数量从字符串转换为浮点数
	cumulativeQuoteQty, err := utils.StrTofloat64(resp.CummulativeQuoteQuantity)
	if err != nil {
		return false, err
	}

	// 获取交易对的基准资产和报价资产
	_, quote, err := AllTradingSymbols.GetBaseAndQuote(resp.Symbol)
	if err != nil {
		return false, fmt.Errorf("failed to get quote asset, err: %s", err.Error())
	}

	// 根据报价资产和累计报价数量计算累计报价价值
	cumulativeQuoteValue, err := Qty2value(quote, cumulativeQuoteQty, bpm)
	if err != nil {
		return false, fmt.Errorf("failed to calculate value, err: %s", err.Error())
	}

	si.L.Lock()
	defer si.L.Unlock()

	// 更新累计报价价值和计划报价价值的差值
	si.CumulativeQuoteValue = si.CumulativeQuoteValue + cumulativeQuoteValue
	si.DeltaQuoteValue = si.PlanQuoteValue - si.CumulativeQuoteValue

	// 返回是否完成交易（计划报价价值小于等于累计报价价值）
	return si.DeltaQuoteValue <= 0, nil
}

// UpdateValue 更新累计报价数量和增量报价数量
func (si *SafeQuoteQtyInfo) UpdateValue(value string) error {
	valueDecimal, err := utils.StrTofloat64(value)
	if err != nil {
		return err
	}

	si.L.Lock()
	defer si.L.Unlock()
	si.CumulativeQuoteValue = si.CumulativeQuoteValue + valueDecimal
	si.DeltaQuoteValue = si.PlanQuoteValue - si.CumulativeQuoteValue

	// succeed
	return nil
}

// UpdateBNBCommission 更新佣金
// return (exit or not, error), the progress will exit if the returned value is true
func (si *SafeQuoteQtyInfo) UpdateBNBCommission(asset, value string) (bool, error) {
	if asset != define.AssetBNB {
		return true, nil
	}

	valueDecimal, err := utils.StrTofloat64(value)
	if err != nil {
		return false, err
	}

	si.L.Lock()
	defer si.L.Unlock()
	if si.Commission == nil {
		si.Commission = make(map[string]float64)
	}

	origin, ok := si.Commission[asset]
	if !ok {
		si.Commission[asset] = origin
	} else {
		si.Commission[asset] = origin + valueDecimal
	}

	return false, nil
}

type QuoteQtyInfo struct {
	PlanQuoteValue       string            `json:"plan_quote_value"`
	CumulativeQuoteValue string            `json:"cumulative_quote_value"`
	DeltaQuoteValue      string            `json:"delta_quote_value"`
	Commission           map[string]string `json:"commission"` // key: asset, value: commission value
}

// SetPlanQuoteqty sets the value of PlanQuoteqty
func (si *SafeQuoteQtyInfo) Get() QuoteQtyInfo {
	si.L.RLock()
	defer si.L.RUnlock()

	commission := make(map[string]string)
	for k, v := range si.Commission {

		commission[k] = utils.Float64ToDecimal(v).StringFixed(define.DecimalPlaces3)
	}

	return QuoteQtyInfo{
		PlanQuoteValue:       utils.Float64ToDecimal(si.PlanQuoteValue).StringFixed(define.DecimalPlaces3),
		CumulativeQuoteValue: utils.Float64ToDecimal(si.CumulativeQuoteValue).StringFixed(define.DecimalPlaces3),
		DeltaQuoteValue:      utils.Float64ToDecimal(si.DeltaQuoteValue).StringFixed(define.DecimalPlaces3),
		Commission:           commission,
	}
}

func (si *SafeQuoteQtyInfo) GetCumulativeQuoteValue() float64 {
	si.L.RLock()
	defer si.L.RUnlock()

	return si.CumulativeQuoteValue
}

func (si *SafeQuoteQtyInfo) GetDeltaQuoteValue() float64 {
	si.L.RLock()
	defer si.L.RUnlock()

	return si.DeltaQuoteValue
}
