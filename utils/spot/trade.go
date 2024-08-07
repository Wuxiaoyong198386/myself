package spot

import (
	"math"

	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/utils"

	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"

	json "github.com/json-iterator/go"
	"github.com/open-binance/logger"
)

type Trader struct {
	TraceID       string `json:"trace_id"`
	Symbol        string `json:"symbol,omitempty"`
	Quantity      string `json:"quantity,omitempty"`
	QuoteOrderQty string `json:"quote_order_qty,omitempty"`
	Price         string `json:"price,omitempty"`
	ClientOrderID string `json:"client_order_id,omitempty"`
	Side          string `json:"side_type,omitempty"`
	Type          string `json:"type"`
	TimeInForce   string `json:"time_in_force,omitempty"`
}

// NewTrader creates a new trader
func NewTrader(symbol, qty, quoteQty, price, traceID string, sideType, orderType string, timeInForce string) *Trader {

	return &Trader{
		TraceID:       traceID,
		Symbol:        symbol,
		Quantity:      qty,
		QuoteOrderQty: quoteQty,
		Price:         price,
		ClientOrderID: define.NewClientOrderID(),
		Side:          sideType,
		Type:          orderType,
		TimeInForce:   timeInForce,
	}
}

func updateQtyWithResp(resp *binance.CreateOrderResponse) {
	_, err := inits.QuoteValueInfo.UpdateValueWithResp(resp, inits.BestPriceInfo)
	if err != nil {
		logger.Infof("msg=failed to update value||err=%s", err.Error())
		return
	}

	// TODO: delete the log
	info := inits.QuoteValueInfo.Get()
	infoJSON, _ := json.Marshal(info)
	logger.Infof("msg=disp quote value info||info=%s", string(infoJSON))

}

/*
 * @description: 创个一个新的交易单，但并没有下单
 * @fileName: trade.go
 * @author: vip120@126.com
 * @date: 2024-03-21 14:29:13
 */

func PriceCorrection(symbol, priceStr string) decimal.Decimal {
	tickSize, ok := inits.SpotPriceFilterInfo.GetTickSizeBySymbol(symbol)
	if !ok {
		logger.Errorf("msg=get no tick size for spot with symbol||symbol=%s", symbol)
		return decimal.Zero
	}

	return correction(priceStr, tickSize)
}

// QuantityCorrection corrects the quantity
func QuantityCorrection(symbol, quantityStr string) decimal.Decimal {
	stepSize, ok := inits.SpotPriceFilterInfo.GetStepSizeBySymbol(symbol)
	if !ok {
		logger.Errorf("msg=get no step size for spot with symbol||symbol=%s", symbol)
		return decimal.Zero
	}

	return correction(quantityStr, stepSize)
}

/*
 * @description: 根据当前价格和过滤器的值算出最小数量
 * @fileName: trade.go
 * @author: vip120@126.com
 * @date: 2024-04-03 16:29:34
 */
func GetMinQuantity(symbol string, currentPrice float64, s string, stepUnit int32) float64 {

	if currentPrice <= 0 {
		logger.Errorf("currentPrice must be greater than 0, got %f", currentPrice)
		return 0
	}
	symbolFilterInfo, ok := inits.SpotPriceFilterInfo.GetBySymbol(symbol)

	if !ok {
		logger.Errorf("get no step size for spot with symbol||symbol=%s", symbol)
		return 0
	}
	minInvestmentValue, err := utils.StrTofloat64(symbolFilterInfo.Notional.MinNotional)
	if err != nil {
		logger.Errorf("Error converting minNotional to float64: %v", err)
		return 0
	}
	minInvestmentQuantity, err := utils.StrTofloat64(symbolFilterInfo.LotSize.MinQty)
	if err != nil {
		logger.Errorf("Error converting minQty to float64: %v", err)
		return 0
	}

	investmentUnitQuantity := calculateInvestmentUnitQuantity(currentPrice, minInvestmentValue, s)

	// 根据步进单位调整投资单元数量
	result := adjustQuantityByStepUnit(investmentUnitQuantity, float64(stepUnit))

	if result < minInvestmentQuantity {
		return minInvestmentQuantity
	}
	return result
}

// 计算投资单元数量的逻辑抽离成一个单独的函数
func calculateInvestmentUnitQuantity(currentPrice float64, minInvestmentValue float64, s string) float64 {
	if s == "s1" {
		return minInvestmentValue * 1.5 / currentPrice
	}
	return minInvestmentValue / currentPrice
}

// 将投资单元数量按照步进单位进行调整的逻辑抽离成一个单独的函数
func adjustQuantityByStepUnit(investmentUnitQuantity float64, stepUnit float64) float64 {
	if stepUnit == 0 {
		return math.Ceil(investmentUnitQuantity)
	}
	return math.Ceil(investmentUnitQuantity/stepUnit) * stepUnit
}

func CalculateMinInvestmentv2(minInvestmentValue, currentPrice, minInvestmentQuantity, stepUnit float64) float64 {
	// 计算投资单元数量
	investmentUnitQuantity := minInvestmentValue / currentPrice
	result := float64(0)

	// 根据步进单位调整投资单元数量
	switch stepUnit {
	case 0:
		// 如果步进单位为1， 则直接取整返回
		result = math.Ceil(investmentUnitQuantity)
	default:
		// 如果步进单位不为 "1"，则按照步进单位进行调整
		result = math.Ceil(investmentUnitQuantity/stepUnit) * stepUnit
	}

	// 如果调整后的投资单元数量小于最小投资数量，则返回最小投资数量
	if result < minInvestmentQuantity {
		return minInvestmentQuantity
	}

	// 返回调整后的投资单元数量
	return result
}

func correction(val string, size string) decimal.Decimal {
	valDecimal, err := utils.NewFromString(val)
	if err != nil {
		logger.Errorf("msg=invalid value||val=%v||err=%s", val, err.Error())
		return decimal.Zero
	}
	return valDecimal.RoundFloor(int32(getTickSizeLength(size)))
}

func getTickSizeLength(tickSize string) int32 {
	var (
		// 长度变量
		length int32 = 0
		// 标记是否开始计算长度
		start = false
		// 字符串变量
		str = tickSize
	)

	// 如果tickSize为空字符串，返回0
	if len(tickSize) == 0 {
		return 0
	}

	// 如果tickSize的第一个字符是'0'
	if tickSize[0] == '0' {
		// 遍历字符串
		for i := range str {
			// 如果遇到字符'1'，跳出循环
			if str[i] == '1' {
				break
			}

			// 如果遇到小数点
			if str[i] == '.' {
				// 标记开始计算长度
				start = true
			}

			// 如果已经开始计算长度
			if start {
				// 长度加1
				length++
			}
		}

		// 返回长度
		return length
	}

	// 遍历字符串
	for i := range str {
		// 如果遇到小数点，跳出循环
		if str[i] == '.' {
			break
		}

		// 如果遇到字符'1'
		if str[i] == '1' {
			// 标记开始计算长度
			start = true
		}

		// 如果已经开始计算长度
		if start {
			// 长度加1
			length++
		}
	}

	// 返回长度减1的相反数
	return -(length - 1)
}
