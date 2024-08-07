package datastruct

import (
	"sync"

	"github.com/shopspring/decimal"
)

type SafeOrderInfo3 struct {
	L            *sync.RWMutex     `json:"l"`
	OrderInfoCid map[string]string `json:"order_info_cid"`
}

func (i SafeOrderInfo3) SetOrderClientOrderID(symbol, clientOrderIDstring string) {
	i.OrderInfoCid[symbol] = clientOrderIDstring
}
func (i SafeOrderInfo3) GetOrderClientOrderID(symbol string) (string, bool) {
	v, ok := i.OrderInfoCid[symbol]
	return v, ok
}
func (i SafeOrderInfo3) DeleteOrderClientOrderID(symbol string) {
	delete(i.OrderInfoCid, symbol)
}

type SafeOrderInfo2 struct {
	L         *sync.RWMutex           `json:"l"`
	OrderInfo map[string][]OrderInfo2 `json:"order_info2"`
}

type OrderInfo2 struct {
	ClientOrderID string          `json:"clientOrderID"` //订单号
	OrdeCode      string          `json:"order_code"`    //订单号后面的编码
	OrderType     string          `json:"order_type"`    //LIMIT,STOP,TAKE_PROFIT
	SideType      string          `json:"side_type"`     //BUY,SELL
	Price         decimal.Decimal `json:"price"`         //价格
	Quantity      decimal.Decimal `json:"quantity"`      //数量
	OrderStatus   string          `json:"order_status"`  //NEW,FILLED,CANCELED
}

func (o *SafeOrderInfo2) SetValue(symbol string, v OrderInfo2) {
	o.L.Lock()
	defer o.L.Unlock()
	for i, v1 := range o.OrderInfo[symbol] {
		if v1.ClientOrderID == v.ClientOrderID {
			o.OrderInfo[symbol][i] = v
			return
		}
	}
	o.OrderInfo[symbol] = append(o.OrderInfo[symbol], v)
}

func (o *SafeOrderInfo2) GetValue(symbol string) ([]OrderInfo2, bool) {
	o.L.RLock()
	defer o.L.RUnlock()
	v, ok := o.OrderInfo[symbol]
	return v, ok
}
func (o *SafeOrderInfo2) DeleteValue(symbol string) {
	o.L.Lock()
	defer o.L.Unlock()
	delete(o.OrderInfo, symbol)
}

type SafeOrderInfo1 struct {
	L         *sync.RWMutex             `json:"l"`
	OrderInfo map[string]OrderLimitInfo `json:"order_info"`
}

type OrderLimitInfo struct {
	Symbol               string          `json:"symbol"`
	ClientOrderID        string          `json:"client_order_id"`
	OrderStopSideType    string          `json:"order_stop_side_type"`
	OrderStopPrice       decimal.Decimal `json:"order_stop_price"`
	OrderStopQuantity    decimal.Decimal `json:"order_stop_quantity"`
	OrderStopAllQuantity decimal.Decimal `json:"order_stop_all_quantity"`
	OrderClientOrderID   string          `json:"order_client_order_id"`
	A2_ok                bool            `json:"a2_ok"`
	A2_stop_price        decimal.Decimal `json:"a2_stop_price"`

	OrderLimitSideType      string          `json:"order_limit_side_type"`
	TakeProfitSideType      string          `json:"take_profit_side_type"`
	TakeProfitK1Close       decimal.Decimal `json:"take_profit_k1_close"`
	TakeProfitK1High        decimal.Decimal `json:"take_profit_k1_high"`
	TakeProfitMa            decimal.Decimal `json:"take_profit_ma"`
	TakeProfitPrice         decimal.Decimal `json:"take_profit_price"`
	TakeProfitQuantity      decimal.Decimal `json:"take_profit_quantity"`
	TakeProfitClientOrderID string          `json:"take_profit_client_order_id"`
}

func (s *SafeOrderInfo1) SetValue(key string, orderinfo *OrderLimitInfo) {
	s.L.Lock()
	defer s.L.Unlock()
	s.OrderInfo[key] = *orderinfo
}
func (s *SafeOrderInfo1) SetValueA2(key string, isOk bool) {
	s.L.Lock()
	defer s.L.Unlock()
	// 假设s.OrderInfo是一个map，key是其键，A2_ok是结构体中的字段
	temp := s.OrderInfo[key] // 取出map中的结构体
	temp.A2_ok = isOk        // 修改结构体的字段
	s.OrderInfo[key] = temp  // 将修改后的结构体放回map中
}

func (s *SafeOrderInfo1) SetValueQtyA1(key string, qty decimal.Decimal) {
	s.L.Lock()
	defer s.L.Unlock()
	temp := s.OrderInfo[key]        // 取出map中的结构体
	temp.OrderStopQuantity = qty    // 止损数量
	temp.TakeProfitQuantity = qty   // 止盈数量
	temp.OrderStopAllQuantity = qty // 总的数量
	s.OrderInfo[key] = temp         // 将修改后的结构体放回map中
}
func (s *SafeOrderInfo1) SetValueQtyA2(key string, qty decimal.Decimal) {
	s.L.Lock()
	defer s.L.Unlock()
	temp := s.OrderInfo[key]                                       // 取出map中的结构体
	temp.OrderStopQuantity = temp.OrderStopQuantity.Add(qty)       // 止损数量
	temp.TakeProfitQuantity = temp.OrderStopQuantity.Add(qty)      // 止盈数量
	temp.OrderStopAllQuantity = temp.OrderStopAllQuantity.Add(qty) // 总的数量
	s.OrderInfo[key] = temp                                        // 将修改后的结构体放回map中
}

func (s *SafeOrderInfo1) GetValueQty(key string) OrderLimitInfo {
	s.L.Lock()
	defer s.L.Unlock()
	return s.OrderInfo[key]
}

func (s *SafeOrderInfo1) GetValueA2(key string) bool {
	s.L.Lock()
	defer s.L.Unlock()
	// 假设s.OrderInfo是一个map，key是其键，A2_ok是结构体中的字段
	return s.OrderInfo[key].A2_ok // 取出map中的结构体
}

func (s *SafeOrderInfo1) GetValue(key string) (OrderLimitInfo, bool) {
	s.L.RLock()
	defer s.L.RUnlock()
	if value, ok := s.OrderInfo[key]; !ok {
		return OrderLimitInfo{}, false
	} else {
		return value, true
	}
}

func (s *SafeOrderInfo1) Delete(key string) {
	s.L.Lock()
	defer s.L.Unlock()
	// 检查CreateOrderTime映射中是否存在该键
	delete(s.OrderInfo, key) // 删除键及其值
}
