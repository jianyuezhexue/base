package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// 本地时间
type LocalTime time.Time

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt.UTC(), nil
}
func (t *LocalTime) Scan(v any) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value.In(time.Local))
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t *LocalTime) String() string {
	// 如果时间 null 那么我们需要把返回的值进行修改
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format("2006-01-02 15:04:05")
}

func (t *LocalTime) ToTime() time.Time {
	return time.Time(*t)
}

// dateString | 年月日
func (t *LocalTime) DateString() string {
	if t == nil || t.IsZero() {
		return ""
	}
	return time.Time(*t).Format("2006-01-02")
}

// IsToday 判断是否为今天
func (t *LocalTime) IsToday() bool {
	return time.Now().In(time.Local).Format("2006-01-02") == t.DateString()
}

// 小于等于今天
func (t *LocalTime) LteToday() bool {
	// 获取当前时间的本地时区零点（今天的结束）
	now := time.Now().Local()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// 将目标时间转换为本地时区，并截取日期部分
	targetLocal := time.Time(*t)
	targetDate := time.Date(targetLocal.Year(), targetLocal.Month(), targetLocal.Day(), 0, 0, 0, 0, targetLocal.Location())

	// 比较日期
	return targetDate.Before(currentDate) || targetDate.Equal(currentDate)
}

func (t *LocalTime) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var err error
	// 前端接收的时间字符串
	str := string(data)
	// 去除接收的str收尾多余的"
	timeStr := strings.Trim(str, "\"")
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	*t = LocalTime(t1)
	return err
}

// string 转 LocalTime
func StringToLocalTime(str string) LocalTime {
	if len(str) == 10 {
		str = str + " 00:00:00"
	}

	t1, _ := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
	return LocalTime(t1)
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(t)
	// 如果时间值是空或者0值 返回为null 如果写空字符串会报错
	if t.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", tTime.Format("2006-01-02 15:04:05")), nil
}

// 字符串数组
type StringArray []string

func (m *StringArray) Scan(val interface{}) error {
	s := val.([]uint8)
	ss := strings.Split(string(s), ",")
	*m = ss
	return nil
}
func (m StringArray) Value() (driver.Value, error) {
	str := strings.Join(m, ",")
	return str, nil
}
