package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/open-binance/logger"
	"github.com/shopspring/decimal"
)

var DB *sql.DB

type Statistics struct {
	OrderId string          `gorm:"column:order_id"` //订单ID
	Symbol  string          `gorm:"column:symbol"`   //交易对
	Side    int             `gorm:"column:side"`     //方向1多2空
	Profit  int             `gorm:"column:profit"`   //盈亏1盈利2亏损
	Money   decimal.Decimal `gorm:"column:money"`    //金额
	Time    string          `gorm:"column:time"`
	Type    string          `gorm:"column:type"` //类型
}

func StatisticsInsert(order_id string, symbol string, side int, profit int, money decimal.Decimal, time string) {
	insData := Statistics{
		Type:    "多仓",
		OrderId: order_id, //  order_id
		Symbol:  symbol,   //  symbol
		Side:    side,
		Profit:  profit,
		Money:   money,
		Time:    time,
	}
	orm, err := InitOrm()
	if err != nil {
		fmt.Println("初始化ORM失败:", err)
	}
	if result := orm.Create(&insData); result.Error != nil {
		fmt.Println("结果添加异常:", result.Error.Error())
	} else {
		fmt.Println("结果添加成功:", insData)
		logger.Infof("结果添加成功%+v", insData)
		StatisticsQuery()
	}
}

func StatisticsQuery() {
	orm, err := InitOrm()
	if err != nil {
		fmt.Println("初始化ORM失败:", err)
	}
	var data []Statistics
	if result := orm.Find(&data); result.Error != nil {
		fmt.Println("查询异常:", result.Error.Error())
	} else {
		var k_kui, k_ying, d_kui, d_ying, kui_count, ying_count int
		var all_money decimal.Decimal
		type ResultStruct struct {
			Kkui   int             `json:"k_kui"`     //空-亏损
			Kying  int             `json:"k_ying"`    //空-盈利
			Dkui   int             `json:"d_kui"`     //多-亏损
			Dying  int             `json:"d_ying"`    //多-盈利
			Kcount int             `json:"k_count"`   //亏损笔数
			Ycount int             `json:"y_count"`   //盈利笔数
			Amoney decimal.Decimal `json:"all_money"` //盈亏金额
		}
		for _, v := range data {
			if v.Side == 1 { //做多
				if v.Profit == 1 {
					d_ying++ //做多盈利笔数
				} else {
					d_kui++ //做多亏损笔数
				}
			} else { //做空
				if v.Profit == 1 {
					k_ying++ //做空盈利笔数
				} else if v.Profit == 2 {
					k_kui++ //做空亏损笔数
				}
			}
			if v.Profit == 1 {
				ying_count++ //盈利笔数
			} else {
				kui_count++ //亏损笔数
			}
			all_money = all_money.Add(v.Money)
		}
		res := ResultStruct{
			Kkui:   k_kui,
			Kying:  k_ying,
			Dkui:   d_kui,
			Dying:  d_ying,
			Kcount: kui_count,
			Ycount: ying_count,
			Amoney: all_money,
		}
		fmt.Printf("统计结果:%+v", res)
		logger.Infof("统计结果%+v", res)
	}
}
