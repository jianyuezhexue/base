package tool

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jianyuezhexue/base/db"
	"github.com/jinzhu/copier"
)

// 拷贝自定义映射关系
var converters = []copier.TypeConverter{
	// 字符串数组转字符串
	{
		SrcType: []string{},
		DstType: copier.String,
		Fn: func(src any) (any, error) {
			s, ok := src.([]string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			return strings.Join(s, ","), nil
		},
	},
	// 字符串转字符串数组
	{
		SrcType: copier.String,
		DstType: []string{},
		Fn: func(src any) (any, error) {
			s, ok := src.(string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			if s == "" {
				return []string{}, nil
			}
			return strings.Split(s, ","), nil
		},
	},
	// 字符串转字符串数组
	{
		SrcType: copier.String,
		DstType: db.StringArray{},
		Fn: func(src any) (any, error) {
			s, ok := src.(string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			if s == "" {
				return []string{}, nil
			}
			return strings.Split(s, ","), nil
		},
	},
	// 字符串数组转字符串
	{
		SrcType: db.StringArray{},
		DstType: copier.String,
		Fn: func(src any) (any, error) {
			s, ok := src.([]string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			return strings.Join(s, ","), nil
		},
	},
	// string 转 db.LocalTime
	{
		SrcType: copier.String,
		DstType: db.LocalTime{},
		Fn: func(src any) (any, error) {
			s, ok := src.(string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			loc, _ := time.LoadLocation("Local")
			t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
			if err != nil {
				return nil, err
			}
			// 关键点：将 time.Time 转换为 db.LocalTime
			return db.LocalTime(t), nil
		},
	},
	// db.LocalTime 转 string
	{
		SrcType: db.LocalTime{},
		DstType: copier.String,
		Fn: func(src any) (any, error) {
			t, ok := src.(time.Time)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			return t.Format("2006-01-02 15:04:05"), nil
		},
	},
	// string 转 int
	{
		SrcType: copier.String,
		DstType: copier.Int,
		Fn: func(src any) (any, error) {
			s, ok := src.(string)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			res, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			return res, nil
		},
	},
	// int 转 string
	{
		SrcType: copier.Int,
		DstType: copier.String,
		Fn: func(src any) (any, error) {
			i, ok := src.(int)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			return strconv.Itoa(i), nil
		},
	},
	{
		SrcType: sql.NullTime{},
		DstType: copier.String,
		Fn: func(src any) (any, error) {
			t, ok := src.(sql.NullTime)
			if !ok {
				return nil, errors.New("src type not matching")
			}
			if !t.Valid {
				return "", nil
			}
			return t.Time.Format("2006-01-02 15:04:05"), nil
		},
	},
	{
		SrcType: copier.String,
		DstType: sql.NullTime{},
		Fn: func(src any) (any, error) {
			s, ok := src.(string)
			if !ok {
				return nil, errors.New("src type not matching")
			}

			res := sql.NullTime{}
			if s == "" {
				res.Valid = false
				return res, nil
			}

			loc, _ := time.LoadLocation("Local")
			time, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
			if err != nil {
				return nil, err
			}
			res.Valid = true
			res.Time = time
			return res, nil
		},
	},
}

// CopyDeep 深度复制结构体
func CopyDeep(target any, source any) error {
	if err := copier.CopyWithOption(target, source, copier.Option{
		DeepCopy:   true,
		Converters: converters,
	}); err != nil {
		return err
	}
	return nil
}

// Copy 浅拷贝
func Copy(target any, source any) error {
	if err := copier.CopyWithOption(target, source, copier.Option{
		Converters: converters,
	}); err != nil {
		return err
	}
	return nil
}
