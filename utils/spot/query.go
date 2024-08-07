package spot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go_code/myselfgo/client"
	"go_code/myselfgo/define"
	"go_code/myselfgo/inits"
	"go_code/myselfgo/modules/datastruct"
	"go_code/myselfgo/utils"

	"github.com/adshao/go-binance/v2"
	json "github.com/json-iterator/go"
	"github.com/open-binance/logger"
)

// FundingInfo specifies funding account info
// reference: https://binance-docs.github.io/apidocs/spot/cn/#user_data-16
type FundingInfo struct {
	Asset        string `json:"asset"`
	Free         string `json:"free"`
	Locked       string `json:"locked"`
	Freeze       string `json:"freeze"`
	Withdrawing  string `json:"withdrawing"`
	BtcValuation string `json:"btcValuation"`
}

// QuerySpotAccountInfo query account info for spot
func QuerySpotAccountInfo(bpm *datastruct.SafeBestPriceMap) (map[string]datastruct.AssetInfo, float64, float64, error) {
	// 获取账户信息
	acount := inits.Config.Account
	apiKey := acount.ApiKey
	secretKey := acount.SecretKey

	// 初始化资产信息映射和总价值
	asset2acount := make(map[string]datastruct.AssetInfo)
	var totalValue float64 = 0
	var totalLockValue float64 = 0

	// 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	defer cancel()

	// 获取账户信息
	res, err := client.GetClient(apiKey, secretKey).NewGetAccountService().Do(ctx)
	if err != nil {
		return asset2acount, totalValue, totalLockValue, fmt.Errorf("binance client 获取账户信息失败, 失败: %s", err.Error())
	}

	if res == nil {
		return asset2acount, totalValue, totalLockValue, errors.New("nil response from binance for spot")
	}

	// 更新时间
	updateTime := time.Now().UnixMilli()

	// 处理资产列表
	//资产列表
	for _, balance := range res.Balances {
		asset := balance.Asset   // 资产类型
		free := balance.Free     // 可提余额
		locked := balance.Locked // 已冻结金额

		// 将字符串类型的可提余额和已冻结金额转换为float64类型
		freeDecimal, err := utils.StrTofloat64(free)
		if err != nil {
			logger.Errorf("msg=%s||free=%s||err=%s",
				"failed to parse free of spot as float64", free, err.Error())
			continue
		}
		lockedDecimal, err := utils.StrTofloat64(locked)
		if err != nil {
			logger.Errorf("msg=%s||locked=%s||err=%s",
				"failed to parse locked of spot as float64", locked, err.Error())
			continue
		}

		// 如果可提余额和已冻结金额都等于0，则跳过该资产
		// 如果可提余额和已冻结金额都等于0
		if freeDecimal == 0 && lockedDecimal == 0 {
			continue
		}

		// 计算可用余额的价值
		freeValue, err := datastruct.Qty2value(asset, freeDecimal, bpm)
		if err != nil {
			// 仅记录日志即可
			// just log is ok
			logger.Errorf("msg=failed to calculate value||asset=%s||err=%s", asset, err.Error())
		}

		// 累加所有可用资产的总价值
		totalValue = totalValue + freeValue
		// 累加所有已冻结资产的总价值
		totalLockValue = totalLockValue + lockedDecimal

		// 将每个资产的账号信息添加到映射中
		//每个资产的账号信息
		asset2acount[asset] = datastruct.AssetInfo{
			Asset:      asset,
			UpdateTime: updateTime,
			Free:       freeDecimal,
			FreeValue:  freeValue,
			Locked:     lockedDecimal,
		}
	}

	return asset2acount, totalValue, totalLockValue, nil
}

// QueryFundingAssetInfo query funding info
func QueryFundingAssetInfo() (map[string]datastruct.AssetInfo, error) {
	asset2acount := make(map[string]datastruct.AssetInfo) // key: asset, value: asset info
	host, _ := inits.NetworkDelayMap.GetHostAndDelay()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(define.TimeoutBinanceAPI)*time.Second)
	defer cancel()

	query := make(map[string]string)
	resp, _, err := CallAPI(ctx, false, http.MethodPost, host, "/sapi/v1/asset/get-funding-asset", query)
	if err != nil {
		logger.Errorf("msg=failed to get funding info||err=%s", err.Error())
		return asset2acount, err
	}

	var fundingInfos []FundingInfo
	if err := json.Unmarshal([]byte(resp), &fundingInfos); err != nil {
		logger.Errorf("msg=failed to unmarshal funding response||err=%s", err.Error())
		return asset2acount, err
	}

	for _, fundingInfo := range fundingInfos {
		asset := fundingInfo.Asset   //币种
		free := fundingInfo.Free     //可用余额
		locked := fundingInfo.Locked //锁定资金
		freeDecimal, err := utils.StrTofloat64(free)
		if err != nil {
			logger.Errorf("msg=%s||asset=%s||free=%s||err=%s",
				"failed to parse free of funding as decimal", asset, free, err.Error())
			continue
		}
		lockedDecimal, err := utils.StrTofloat64(locked)
		if err != nil {
			logger.Errorf("msg=%s||asset=%s||locked=%s||err=%s",
				"failed to parse locked of funding as decimal", asset, locked, err.Error())
			continue
		}
		//如果可用余额 <=0 且 锁定资金 <=0 就跳过
		if freeDecimal <= 0 && lockedDecimal <= 0 {
			continue
		}
		//否侧就返回资金账户
		asset2acount[asset] = datastruct.AssetInfo{
			Asset:  asset,
			Free:   freeDecimal,
			Locked: lockedDecimal,
		}
	}

	return asset2acount, nil
}

func GetAssetDividend() (*binance.DividendResponseWrapper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()

	client := client.GetClient(inits.Config.Account.ApiKey, inits.Config.Account.SecretKey).NewAssetDividendService().Limit(500)

	resp, err := client.Do(ctx)
	return resp, err
}

func QueryCommission(symbol, traceID string) (CommissionInfo, error) {
	var commission CommissionInfo
	host, _ := inits.NetworkDelayMap.GetHostAndDelay()
	cost, body, _, err := DoQueryCommission(host, symbol)
	if err != nil {
		return commission, err
	}
	logger.Infof("msg=%s||host=%s||symbol=%s||cost=%.3fms||trace_id=%s",
		"succeed to query commission", host, symbol, cost, traceID)

	if err := json.Unmarshal(body, &commission); err != nil {
		return commission, fmt.Errorf("failed to unmarshal, err: %s", err.Error())
	}

	return commission, nil
}

/*
 * @description: 获取当前账户的佣金费率
 * @fileName: query.go
 * @author: vip120@126.com
 * @date: 2024-03-20 11:52:25
 */
func DoQueryCommission(host, symbol string) (float64, []byte, http.Header, error) {
	params := map[string]string{
		"symbol": symbol,
	}
	ctx, cancel := context.WithTimeout(context.Background(), define.TimeoutBinanceAPI)
	defer cancel()
	start := time.Now()
	res, header, err := CallAPI(ctx, false, http.MethodGet, host, "/api/v3/account/commission", params)
	cost := time.Since(start).Seconds() * 1000 // unit: ms

	return cost, res, header, err
}
