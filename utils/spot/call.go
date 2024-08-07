package spot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	json "github.com/json-iterator/go"
)

// UpdateMsgFromHeader updates the message received from binance
func UpdateMsgFromHeader(header http.Header) {
	// update 1m used weight by the way
	inits.UsedWeight1m.UpdateUsedWeight1m(header)
	// update 10s order count by the way
	inits.OrderCount10s.UpdateOrderCount10s(header)
	// update 1d order count by the way
	inits.OrderCount1d.UpdateOrderCount1d(header)
}

// Ping calls the binance api named "/api/v3/ping", and get cost, response body and header
func Ping(host string) (float64, []byte, http.Header, error) {
	params := make(map[string]string)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	defer cancel()
	start := time.Now()
	res, header, err := CallAPI(ctx, true, http.MethodGet, host, "/api/v3/ping", params)
	cost := time.Since(start).Seconds() * 1000 // unit: ms

	return cost, res, header, err
}

// CreateSpotOrder creates spot order
// return (host, cost, resp, header, error)
func CreateSpotOrder(params map[string]string) (string, float64, *binance.CreateOrderResponse, http.Header, error) {
	host, _ := inits.NetworkDelayMap.GetHostAndDelay()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	defer cancel()
	var resp *binance.CreateOrderResponse

	start := time.Now()
	resBody, header, err := CallAPI(ctx, false, http.MethodPost, host, "/api/v3/order", params)
	cost := time.Since(start).Seconds() * 1000 // unit: ms
	if err != nil {
		return host, cost, resp, header, err
	}

	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		return host, cost, resp, header, fmt.Errorf("failed to unmarshal, err: %s", err.Error())
	}

	// succeed
	return host, cost, resp, header, nil
}

// CallAPI calls spot api
func CallAPI(ctx context.Context, ping bool, method, host, path string, params map[string]string) ([]byte, http.Header, error) {
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey
	client := client.GetClient(apiKey, secretKey)
	if !ping {
		params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	requestURL := signatureURL(ping, host, path, client.SecretKey, params)

	var header http.Header
	req, err := http.NewRequest(method, requestURL.String(), nil)
	if err != nil {
		return []byte{}, header, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("X-MBX-APIKEY", client.APIKey)

	rawResponse, err := client.HTTPClient.Do(req)
	if err != nil {
		return []byte{}, header, err
	}
	defer rawResponse.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		return []byte{}, header, err
	}
	header = rawResponse.Header

	// update message by the way
	go UpdateMsgFromHeader(header)

	if rawResponse.StatusCode >= http.StatusBadRequest {
		apiErr := new(common.APIError)
		if err := json.Unmarshal(bodyBytes, apiErr); err != nil {
			return bodyBytes, header, fmt.Errorf("failed to unmarshal, err: %s", err.Error())
		}
		return bodyBytes, header, apiErr
	}

	return bodyBytes, header, nil
}

func signatureURL(ping bool, host, path, secret string, params map[string]string) url.URL {
	requestURL := url.URL{}
	requestURL.Scheme = "https"
	requestURL.Host = host
	requestURL.Path = path

	query := requestURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}

	requestURL.RawQuery = query.Encode()
	if !ping {
		query.Set("signature", signature(requestURL.RawQuery, secret))
	}
	requestURL.RawQuery = query.Encode()

	return requestURL
}

func signature(src, secret string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
