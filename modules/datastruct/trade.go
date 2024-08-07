package datastruct

type WsOrderUpdate struct {
	Symbol                  string          `json:"s"` //交易对
	ClientOrderId           string          `json:"c"` //客户端订类ID
	Side                    string          `json:"S"` //买卖方向
	Type                    string          `json:"o"` //订单类型
	TimeInForce             TimeInForceType `json:"f"` //订单有效方式
	Volume                  string          `json:"q"` //订单原始数量
	Price                   string          `json:"p"` //订单原始价格
	StopPrice               string          `json:"P"` //止盈止损单触发价格
	IceBergVolume           string          `json:"F"` //冰山订单数量
	OrderListId             int64           `json:"g"` //OCO订单 OrderListId
	OrigCustomOrderId       string          `json:"C"` //原始订单自定义ID(原始订单，指撤单操作的对象。撤单本身被视为另一个订单)
	ExecutionType           string          `json:"x"` //本次事件的具体执行类型 NEW/TRADE...
	Status                  string          `json:"X"` //订单的当前状态
	RejectReason            string          `json:"r"` //订单被拒绝的原因
	Id                      int64           `json:"i"` //order id
	LatestVolume            string          `json:"l"` //订单末次成交量
	FilledVolume            string          `json:"z"` //订单累计已成交量
	LatestPrice             string          `json:"L"` //订单末次成交价格
	FeeAsset                string          `json:"N"` //手续费资产类别
	FeeCost                 string          `json:"n"` //手续费数量
	TransactionTime         int64           `json:"T"` //成交时间
	TradeId                 int64           `json:"t"` //成交ID
	IgnoreI                 int64           `json:"I"` //ignore
	IsInOrderBook           bool            `json:"w"` //is the order in the order book?
	IsMaker                 bool            `json:"m"` //is this order maker?
	IgnoreM                 bool            `json:"M"` //ignore
	CreateTime              int64           `json:"O"` //订单创建时间
	FilledQuoteVolume       string          `json:"Z"` //订单累计已成交金额
	LatestQuoteVolume       string          `json:"Y"` //订单末次成交金额
	QuoteVolume             string          `json:"Q"` //报价数量
	SelfTradePreventionMode string          `json:"V"`

	//这些字段只有在满足某些条件时才会出现在有效载荷中
	TrailingDelta              int64  `json:"d"` // Appears only for trailing stop orders.
	TrailingTime               int64  `json:"D"`
	StrategyId                 int64  `json:"j"` // Appears only if the strategyId parameter was provided upon order placement.
	StrategyType               int64  `json:"J"` // Appears only if the strategyType parameter was provided upon order placement.
	PreventedMatchId           int64  `json:"v"` // Appears only for orders that expired due to STP.
	PreventedQuantity          string `json:"A"`
	LastPreventedQuantity      string `json:"B"`
	TradeGroupId               int64  `json:"u"`
	CounterOrderId             int64  `json:"U"`
	CounterSymbol              string `json:"Cs"`
	PreventedExecutionQuantity string `json:"pl"`
	PreventedExecutionPrice    string `json:"pL"`
	PreventedExecutionQuoteQty string `json:"pY"`
	WorkingTime                int64  `json:"W"` // Appears when the order is working on the book
	MatchType                  string `json:"b"`
	AllocationId               int64  `json:"a"`
	WorkingFloor               string `json:"k"`  // Appears for orders that could potentially have allocations
	UsedSor                    bool   `json:"uS"` // Appears for orders that used SOR
}
