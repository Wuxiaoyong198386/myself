package spot

type CommissionInfo struct {
	Symbol             string                 `json:"symbol,omitempty"`
	StandardCommission StandardCommissionInfo `json:"standardCommission,omitempty"`
	TaxCommission      TaxCommissionInfo      `json:"taxCommission,omitempty"`
	Discount           DiscountInfo           `json:"discount,omitempty"`
}

type StandardCommissionInfo struct {
	Maker  string `json:"maker,omitempty"`
	Taker  string `json:"taker,omitempty"`
	Buyer  string `json:"buyer,omitempty"`
	Seller string `json:"seller,omitempty"`
}

type TaxCommissionInfo struct {
	Maker  string `json:"maker,omitempty"`
	Taker  string `json:"taker,omitempty"`
	Buyer  string `json:"buyer,omitempty"`
	Seller string `json:"seller,omitempty"`
}

type DiscountInfo struct {
	EnabledForAccount bool   `json:"enabledForAccount,omitempty"`
	EnabledForSymbol  bool   `json:"enabledForSymbol,omitempty"`
	DiscountAsset     string `json:"discountAsset,omitempty"`
	Discount          string `json:"discount,omitempty"`
}
