package utils

import (
	"bytes"
	"fmt"
	"go_code/myselfgo/define"

	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/shopspring/decimal"
)

func SetProxy(proxy string) error {
	if err := os.Setenv("HTTP_PROXY", proxy); err != nil {
		return fmt.Errorf("http代理失败，失败原因:%s", err.Error())
	}

	if err := os.Setenv("HTTPS_PROXY", proxy); err != nil {
		return fmt.Errorf("https代理失败，失败原因:%s", err.Error())
	}
	return nil
}

// JoinStrWithSep joins multi string with the separator
// @sep: such as: "-", "&", "_", and so on
// @originStr: the string to be joined
// return: (string, error), for example: ("a=1&b=2", nil)
func JoinStrWithSep(sep string, originStr ...string) string {
	size := len(originStr)
	if size == 0 {
		return ""
	}
	var buf bytes.Buffer
	var i int
	for i = 0; i < size-1; i++ {
		buf.WriteString(originStr[i])
		buf.WriteString(sep)
	}
	buf.WriteString(originStr[i])
	return buf.String()
}

// NewFromString returns a new Decimal from a string representation safely
func NewFromString(value string) (decimal.Decimal, error) {
	return decimal.NewFromString(value)
}

// NewFromFloat converts a float64 to Decimal safely
func NewFromFloat(value float64) (decimal.Decimal, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return decimal.Zero, fmt.Errorf("cannot create a Decimal from %v", value)
	}
	return decimal.NewFromFloat(value), nil
}

// NewFromFloat32 converts a float32 to Decimal safely
func NewFromFloat32(value float32) (decimal.Decimal, error) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return decimal.Zero, fmt.Errorf("cannot create a Decimal from %v", value)
	}
	return decimal.NewFromFloat32(value), nil
}

/*
 * @description: 字符串转 fload 64
 * @fileName: function.go
 * @author: vip120@126.com
 * @date: 2024-04-02 17:03:27
 */
func StrTofloat64(str string) (float64, error) {
	floatValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Println("转换错误:", err)
		return floatValue, err
	}

	// 格式化为8位小数
	formattedValue := fmt.Sprintf("%.8f", floatValue)
	return StrTofloat64_8(&formattedValue)
}

func Float64ToString(float64Value float64) string {
	return fmt.Sprintf("%f", float64Value)
}
func StrTofloat64_8(str *string) (float64, error) {
	floatValue, err := strconv.ParseFloat(*str, 64)
	if err != nil {
		fmt.Println("转换错误:", err)
		return floatValue, err
	}
	return floatValue, err
}

func DecimalToFloat(dec decimal.Decimal) float64 {
	f := new(big.Float)
	f.SetString(dec.String()) // 将decimal转换为big.Float
	f64, _ := f.Float64()     // 将big.Float转换为float64
	return f64
}

func StringFixed(flo *float64) string {
	// 使用fmt.Sprintf来格式化浮点数，保留两位小数
	str := fmt.Sprintf("%.3f", *flo)
	return str
}

/*
 * @description: 字符串转Decimal
 * @fileName: function.go
 * @author: vip120@126.com
 * @date: 2024-04-02 17:02:01
 */
func StrToDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}

func FormatTime() string {
	now := time.Now()
	formatted := now.Format("2006-01-02 15:04:05")
	return formatted
}

func RoundDecimal(d decimal.Decimal, n int) decimal.Decimal {

	// 使用Round方法进行四舍五入
	rounded := d.Round(int32(n))
	return rounded
}

func RoundDecimal45v2(d decimal.Decimal, n int) decimal.Decimal {
	n32 := int32(n)
	r, _ := decimal.NewFromString(d.Round(n32).StringFixed(n32))
	return r
}

func RoundDecimal45(d decimal.Decimal, n int) decimal.Decimal {

	// 检查n是否为负数，因为负数表示要截断到整数位之前
	if n < 0 {
		// 这里为了简化，我们只处理截断到小数点后的情况，如果需要处理到整数位之前，
		// 可以扩展这个函数来处理这种情况
		return decimal.Zero
	}

	// 计算10的n次方
	powerOfTen := decimal.NewFromFloat(math.Pow10(n))

	// 乘以10的n次方
	multiplied := d.Mul(powerOfTen)

	// 截断整数部分（实际上是向下取整）
	truncated := multiplied.Floor().IntPart()

	// 将截断后的整数部分转回decimal并除以10的n次方
	return decimal.NewFromInt(truncated).Div(powerOfTen)
}

func CountDecimalPlaces(f string) int {
	// 将浮点数转换为字符串
	df, _ := decimal.NewFromString(f)
	// 查找小数点的位置
	dfstring := df.String()
	decimalIdx := strings.IndexByte(dfstring, '.')
	if decimalIdx == -1 {
		// 没有小数点，返回0
		return 0
	}
	parts := strings.Split(dfstring, ".")
	return len(parts[1])
}

/*
 * @description: float64转Decimal
 * @fileName: function.go
 * @author: vip120@126.com
 * @date: 2024-04-02 17:02:29
 */
func Float64ToDecimal(f float64) decimal.Decimal {
	d := decimal.NewFromFloat(f)
	return d
}

func FormatDataPrice(data float64, step string) decimal.Decimal {
	step1 := StrToDecimal(step)
	data1 := Float64ToDecimal(data)
	return data1.Mul(step1).Div(step1)
}

func FormatDataQty(data decimal.Decimal, step string) decimal.Decimal {
	step1 := StrToDecimal(step)
	return data.Mul(step1).Div(step1)
}

// GetBaseAndQuoteAsset 获取交易对的基础数据和报价信息
// return (基本资产、报价资产、错误)
// HasSuffix 测试字符串symbol是否以后缀字符串结符
func GetBaseAndQuoteAsset(symbol string) (string, string, error) {
	quoteAsset := ""
	if strings.HasSuffix(symbol, define.QuoteUSDT) {
		quoteAsset = define.QuoteUSDT
	} else if strings.HasSuffix(symbol, define.QuoteTUSD) {
		quoteAsset = define.QuoteTUSD
	} else if strings.HasSuffix(symbol, define.QuoteBUSD) {
		quoteAsset = define.QuoteBUSD
	} else if strings.HasSuffix(symbol, define.QuoteBTC) {
		quoteAsset = define.QuoteBTC
	} else if strings.HasSuffix(symbol, define.QuoteETH) {
		quoteAsset = define.QuoteETH
	} else if strings.HasSuffix(symbol, define.QuoteFDUSD) {
		quoteAsset = define.QuoteFDUSD
	} else if strings.HasSuffix(symbol, define.QuoteUSDC) {
		quoteAsset = define.QuoteUSDC
	} else if strings.HasSuffix(symbol, define.QuoteEUR) {
		quoteAsset = define.QuoteEUR
	} else if strings.HasSuffix(symbol, define.QuoteBNB) {
		quoteAsset = define.QuoteBNB
	} else if strings.HasSuffix(symbol, define.QuoteTRY) {
		quoteAsset = define.QuoteTRY
	} else {
		return "", "", fmt.Errorf("unknown quote asset of the symbol: '%s'", symbol)
	}

	baseAsset := symbol[0 : len(symbol)-len(quoteAsset)]

	return baseAsset, quoteAsset, nil
}
func IsStableCoin(quote string) bool {
	return quote == define.QuoteUSDT
}

func IsSpecialCoin(coin string) bool {
	return coin == define.CoinTRY
}

func Asset2Symbol(asset string) (string, bool) {
	isSpecial := IsSpecialCoin(asset)
	if isSpecial {
		return JoinStrWithSep("", define.CoinUSDT, asset), isSpecial
	}

	return JoinStrWithSep("", asset, define.CoinUSDT), isSpecial
}

func GenUniqueID() string {
	guid := xid.New()
	return guid.String()
}
func MinDecimal(nums ...float64) float64 {
	size := len(nums)
	if size == 0 {
		return define.Float0
	}
	min := nums[0]

	for i := 1; i < size; i++ {
		if nums[i] < min {
			min = nums[i]
		}
	}

	return min
}

func ClientOrderId(sideType string) string {
	clientOrderID := ""
	sep := define.SepEmpty
	suffix := fmt.Sprintf("%d", time.Now().UnixMicro())
	if sideType == define.SideTypeBuy {
		clientOrderID = JoinStrWithSep(sep, "Buy", suffix)
	} else if sideType == define.SideTypeSell {
		clientOrderID = JoinStrWithSep(sep, "Sell", suffix)
	} else {
		clientOrderID = JoinStrWithSep(sep, "Unknown", suffix)
	}
	return clientOrderID
}

// ProfitLossRatio 计算盈亏比并返回结果
// 如果orderPrice等于bidPrice，则返回0
func ProfitLossRatio(orderPrice, bestPrice float64) (float64, string) {
	var profitLossRatio float64
	if orderPrice > bestPrice {
		//计算盈利比
		profitLossRatio = (orderPrice - bestPrice) / orderPrice
		return profitLossRatio, "profit"
	} else if orderPrice < bestPrice {
		//计算亏损比
		profitLossRatio = (bestPrice - orderPrice) / bestPrice
		return profitLossRatio, "loss"
	} else {
		profitLossRatio = define.Float0
		return profitLossRatio, ""
	}
}
