package skit

import (
	"fmt"
	"strings"
	"time"
)

//当前日期
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

/*
time包对于时间的详细定义(记忆：6--1-2-3-4-5 ==> 2006-01-02 15:04:05)

  月份 1,01,Jan,January
  日　 2,02,_2
  时　 3,03,15,PM,pm,AM,am
  分　 4,04
  秒　 5,05
  年　 06,2006
  时区 -07,-0700,Z0700,Z07:00,-07:00,MST
  周几 Mon,Monday

比如小时的表示(原定义是下午3时，也就是15时)
  3 用12小时制表示，去掉前导0
  03 用12小时制表示，保留前导0
  15 用24小时制表示，保留前导0
  03pm 用24小时制am/pm表示上下午表示，保留前导0
  3pm 用24小时制am/pm表示上下午表示，去掉前导0
又比如月份
  1 数字表示月份，去掉前导0
  01 数字表示月份，保留前导0
  Jan 缩写单词表示月份
  January 全单词表示月份
*/
//格式化日期时间
func FormatDateTime(format string, ts ...time.Time) string {
	patterns := []string{
		// 年
		"YYYY", "2006", // 4 位数字完整表示的年份
		"yyyy", "2006", // 4 位数字完整表示的年份
		"YY", "06", // 2 位数字表示的年份
		"yy", "06", // 2 位数字表示的年份

		// 月
		"mmmm", "January", // 月份，完整的文本格式，例如 January 或者 March
		"mmm", "Jan", // 三个字母缩写表示的月份

		"MM", "01", // 数字表示的月份，有前导零
		"mm", "01", // 数字表示的月份，有前导零
		"M", "1", // 数字表示的月份，没有前导零
		"m", "1", // 数字表示的月份，没有前导零

		// 日
		"dddd", "Monday", // 星期几，完整的文本格式;L的小写字母
		"ddd", "Mon", // 星期几，文本表示，3 个字母

		"DD", "02", // 月份中的第几天，有前导零的 2 位数字
		"dd", "02", // 月份中的第几天，有前导零的 2 位数字
		"D", "2", // 月份中的第几天，没有前导零
		"d", "2", // 月份中的第几天，没有前导零

		// 时间
		"HH", "15", // 小时，24 小时格式，有前导零
		"hh", "03", // 小时，12 小时格式，有前导零
		"H", "3", // 小时，12 小时格式，没有前导零
		"h", "3", // 小时，12 小时格式，没有前导零

		"NN", "04", // 有前导零的分钟数
		"nn", "04", // 有前导零的分钟数
		"N", "4", // 没有前导零的分钟数
		"n", "4", // 没有前导零的分钟数
		"SS", "05", // 秒数，有前导零
		"ss", "05", // 秒数，有前导零
		"S", "5", // 秒数，没有前导零
		"s", "5", // 秒数，没有前导零
		"zzz", "000", //毫秒 ".000" 有前导零
		"ZZZ", "000", //毫秒
		"z", "999", //毫秒 ".999" 没有前导零
		"Z", "999", //毫秒

		"a", "pm", // 小写的上午和下午值
		"A", "PM", // 小写的上午和下午值

	}
	replacer := strings.NewReplacer(patterns...)
	format = replacer.Replace(format)

	t := time.Now()
	if len(ts) > 0 {
		t = ts[0]
	}
	return t.Format(format)
}

//字符串转换为时间
func StrToDateTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		"2006-01-02 15:04:05",      //记忆：6--1-2-3-4-5
		"2006-01-02 15:04:05.999Z", //MS-SQL
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 -0700",
		"2006/01/02 15:04:05 -0700 MST",
		"2006/01/02 15:04:05 -0700",
		"2006/01/02 15:04:05",
		"2006-01-02 -0700 MST",
		"2006-01-02 -0700",
		"2006-01-02",
		"2006/01/02 -0700 MST",
		"2006/01/02 -0700",
		"2006/01/02",
		"2006-01-02 15:04:05 -0700 -0700",
		"2006/01/02 15:04:05 -0700 -0700",
		"2006-01-02 -0700 -0700",
		"2006/01/02 -0700 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, value)
		if err == nil {
			return t
		}
	}
	panic(err)
}

func StrToLocalTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	zoneName, offset := time.Now().Zone()

	zoneValue := offset / 3600 * 100
	if zoneValue > 0 {
		value += fmt.Sprintf(" +%04d", zoneValue)
	} else {
		value += fmt.Sprintf(" -%04d", zoneValue)
	}

	if zoneName != "" {
		value += " " + zoneName
	}
	return StrToDateTime(value)
}

//日期时间转为字符串
func DatetimeToStr(d time.Time) string {
	return d.Format("2006-01-02 15:04:05")
}

//日期转为字符串
func DateToStr(d time.Time) string {
	return d.Format("2006-01-02")
}

//时间转为字符串
func TimeToStr(d time.Time) string {
	return d.Format("15:04:05")
}

// 获取当前时间戳
func NowTimestamp() int64 {
	return time.Now().Unix() //时间戳（秒）
	//return time.Now().UnixNano()       //时间戳（纳秒）
	//return time.Now().UnixNano() / 1e6 //时间戳（毫秒）
	//return time.Now().UnixNano() / 1e9 //时间戳（纳秒转换为秒）
}

// 时间戳转化为当前中国时间 GMT，返回string
// 自己指定日期格式 例如 2006-01-02 15:04:05
func TimestampToChina(timestamp int64, format string) string {
	return time.Unix(timestamp, 0).Format(format)
}

// 中国时间字符串转化为时间戳
func ChinaToTimestamp(tmChina string, format string) int64 {
	tm, _ := time.Parse(format, tmChina)
	ts := tm.Unix() - 3600*8
	return ts
}

//2个时间(t1-t2)相差的天数
func TimeSubDay(t1, t2 time.Time) int {
	t1 = t1.UTC().Truncate(24 * time.Hour)
	t2 = t2.UTC().Truncate(24 * time.Hour)
	return int(t1.Sub(t2).Hours() / 24)
}

//时间加减durationString（ns,us,ms,s,m,h） 加1天: 24h 减1天: -24h; 299ms, -1.5h, 2h40m
func TimeAdd(t time.Time, duration string, iMul ...time.Duration) time.Time {
	h, err := time.ParseDuration(duration)
	if err != nil {
		return t
	}
	if len(iMul) > 0 {
		h *= iMul[0]
	}
	return t.Add(h)
}
