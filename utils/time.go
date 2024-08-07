package utils

import (
	"time"
)

const (
	datetimeFormat string = "2006-01-02 15:04:05"
)

func IsExpired(expireTime string) (bool, error) {
	endTime, err := time.Parse(datetimeFormat, expireTime)
	if err != nil {
		return true, err
	}
	return endTime.Before(time.Now()), nil
}

func GetChinaStandTime() time.Time {
	now := time.Now()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return now
	}
	return now.In(loc)
}

func GetChinaStandTimeStr() string {
	now := time.Now()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "unknown location"
	}
	return now.In(loc).Format(datetimeFormat)
}

// TimestampMs2DatetimeStr converts timestamp to datetime string
// @timestamp: timestamp, unit is second, eg: 1652102543000
// return datetime string, eg: "2022-05-09 21:22:23"
func TimestampMs2DatetimeStr(timestamp int64) string {
	return time.UnixMilli(timestamp).Format(datetimeFormat)
}
