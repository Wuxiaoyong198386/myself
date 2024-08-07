package sqlite

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

type Warning struct {
	OrderId   string          `gorm:"column:order_id"`      //订单ID
	Symbol    string          `gorm:"column:symbol"`        //交易对
	Side      int             `gorm:"column:side"`          //方向1多2空
	Warning   decimal.Decimal `gorm:"column:warning_price"` //预警价格
	UpKline   decimal.Decimal `gorm:"column:up_kline"`      //上一条K线价格
	ThisKline decimal.Decimal `gorm:"column:this_kline"`    //这次K线价格
	Boll      decimal.Decimal `gorm:"column:boll"`          //boll差
	Time      string          `gorm:"column:time"`
	Type      string          `gorm:"column:type"` //类型
}

func WarngingInsert(order_id string, symbol string, side int, warning decimal.Decimal, up_kline decimal.Decimal, this_kline decimal.Decimal, boll decimal.Decimal, time string) {
	insData := Warning{
		Type:      "多仓",
		OrderId:   order_id, //  order_id
		Symbol:    symbol,   //  symbol
		Side:      side,
		Warning:   warning,
		UpKline:   up_kline,
		ThisKline: this_kline,
		Boll:      boll,
		Time:      time,
	}
	orm, err := InitOrm()
	if err != nil {
		fmt.Println("初始化ORM失败:", err)
	}
	if result := orm.Create(&insData); result.Error != nil {
		fmt.Println("预警添加异常:", result.Error.Error())
	} else {
		fmt.Println("预警添加成功:", insData)
		logger.Infof("预警添加成功%+v", insData)
	}
}
