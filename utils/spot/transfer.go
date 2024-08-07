package spot

import (
	"context"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"

	"github.com/adshao/go-binance/v2"
)

// Transfer specifies the info for funding transfer
type Transfer struct {
	apiKey    string
	secretKey string
	Asset     string
	Amount    float64
}

// NewTransfer creates a new transfer
func NewTransfer(apiKey, secretKey, asset string, amount float64) (*Transfer, error) {
	transfer := &Transfer{
		apiKey:    apiKey,
		secretKey: secretKey,
		Asset:     asset,
		Amount:    amount,
	}

	return transfer, nil
}

// transfer from funding to main
func (t *Transfer) Funding2Main() (*binance.CreateUserUniversalTransferResponse, error) {
	return t.doTransfer("FUNDING_MAIN")
}

func (t *Transfer) doTransfer(transferType string) (*binance.CreateUserUniversalTransferResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()

	res, err := client.GetClient(t.apiKey, t.secretKey).NewUserUniversalTransferService().
		Type(transferType).Asset(t.Asset).Amount(t.Amount).Do(ctx)

	return res, err
}
