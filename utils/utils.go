package utils

import (
	"time"
)

func StringToTime(s string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	return t, err
}

//将"2016-04-22T21:47:49.694123232+08:00"或者"2016-04-22T21:47:49+08:00"等格式转化为time.Time
func StringToTime1(s string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("2006-01-02T15:04:05+08:00", s, loc)
	return t, err
}
