package sqlite

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/open-binance/logger"
)

type Market struct {
	StartTime string  `gorm:"column:start_time"` //开始时间
	EndTime   string  `gorm:"column:end_time"`   //结束时间
	Up        int     `gorm:"column:up"`         //上涨
	Down      int     `gorm:"column:down"`       //下跌
	UpRate    float64 `gorm:"column:up_rate"`
	DownRate  float64 `gorm:"column:down_rate"`
}

func MarketInsert(start_time, end_time string, up, down int, up_rate, down_rate float64) {
	insData := Market{
		StartTime: start_time, //  order_id
		EndTime:   end_time,   //  symbol
		Up:        up,
		Down:      down,
		UpRate:    up_rate,
		DownRate:  down_rate,
	}
	orm, err := InitOrm()
	if err != nil {
		logger.Error("初始化ORM失败:" + err.Error())
	}
	if result := orm.Create(&insData); result.Error != nil {
		logger.Error("市场情绪添加异常:" + result.Error.Error())
	} else {
		logger.Infof("市场情绪添加成功%+v", insData)
	}
}

func MarketQuery() Market {
	orm, err := InitOrm()
	if err != nil {
		fmt.Println("初始化ORM失败:", err)
	}
	var data Market
	if result := orm.Order("id DESC").First(&data); result.Error != nil {
		logger.Error("市场情绪查询异常" + result.Error.Error())
	}
	logger.Infof("市场情绪查询结果%+v", data)
	return data
}
