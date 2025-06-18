package db

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Condition interface {
	SetWhere(k string, v []interface{})
	SetOr(k string, v []interface{})
	SetOrder(k string)
	SetJoinOn(t, on string) Condition
	SetPage(k string)
	SetPageSize(k string)
}

type GormCondition struct {
	GormPublic
	Join     []*GormJoin
	Page     string
	PageSize string
}

type GormPublic struct {
	Where map[string][]interface{}
	Order []string
	Or    map[string][]interface{}
}

type GormJoin struct {
	Type   string
	JoinOn string
	GormPublic
}

func (e *GormJoin) SetJoinOn(t, on string) Condition {
	return nil
}
func (e *GormJoin) SetPage(k string) {
}
func (e *GormJoin) SetPageSize(k string) {
}

func (e *GormPublic) SetWhere(k string, v []interface{}) {
	if e.Where == nil {
		e.Where = make(map[string][]interface{})
	}
	e.Where[k] = v
}

func (e *GormPublic) SetOr(k string, v []interface{}) {
	if e.Or == nil {
		e.Or = make(map[string][]interface{})
	}
	e.Or[k] = v
}

func (e *GormPublic) SetOrder(k string) {
	if e.Order == nil {
		e.Order = make([]string, 0)
	}
	e.Order = append(e.Order, k)
}

func (e *GormCondition) SetJoinOn(t, on string) Condition {
	if e.Join == nil {
		e.Join = make([]*GormJoin, 0)
	}
	join := &GormJoin{
		Type:       t,
		JoinOn:     on,
		GormPublic: GormPublic{},
	}
	e.Join = append(e.Join, join)
	return join
}

func (e *GormCondition) SetPage(k string) {
	if e.Page != "" {
		return
	}
	if k == "" {
		return
	}
	e.Page = k
}

func (e *GormCondition) SetPageSize(k string) {
	if e.PageSize != "" {
		return
	}
	if k == "" {
		return
	}
	e.PageSize = k
}

type resolveSearchTag struct {
	Type   string   // 条件类型
	Column string   // 表字段
	Table  string   // 数据表
	On     []string // 关联条件[关联表字段,原表字段]
	Join   string   // 关联表
}

// makeTag 解析search的tag标签
func makeTag(tag string) *resolveSearchTag {
	r := &resolveSearchTag{}
	tags := strings.Split(tag, ";")
	var ts []string
	for _, t := range tags {
		ts = strings.Split(t, ":")
		if len(ts) == 0 {
			continue
		}
		switch ts[0] {
		case "type":
			if len(ts) > 1 {
				r.Type = ts[1]
			}
		case "column":
			if len(ts) > 1 {
				r.Column = ts[1]
			}
		case "table":
			if len(ts) > 1 {
				r.Table = ts[1]
			}
		case "on":
			if len(ts) > 1 {
				r.On = ts[1:]
			}
		case "join":
			if len(ts) > 1 {
				r.Join = ts[1]
			}
		case "page":
			r.Type = "page"
		case "pageSize":
			r.Type = "pageSize"
		}
	}
	return r
}

const (
	// FromQueryTag tag标记
	FromQueryTag = "search"
	// Mysql 数据库标识
	Mysql = "mysql"
	// Postgres 数据库标识
	Postgres = "postgres"
)

// ResolveSearchQuery 解析
/**
 * 	eq / exact / iexact 等于
 * 	like / contains / icontains 包含
 *	gt / gte 大于 / 大于等于
 *	lt / lte 小于 / 小于等于
 *	startswith / istartswith 以…起始
 *	endswith / iendswith 以…结束
 *	between 介于…之间
 *	in 存在于...数组
 *	isnull
 *  order 排序		e.g. order[key]=desc     order[key]=asc
 */
func ResolveSearchQuery(driver string, q interface{}, condition Condition) {
	qType := reflect.TypeOf(q)
	if qType.Kind() != reflect.Struct { // 跳过非结构体类型
		return
	}
	qValue := reflect.ValueOf(q)
	var tag string
	var ok bool
	var t *resolveSearchTag
	for i := 0; i < qType.NumField(); i++ {
		tag, ok = "", false
		tag, ok = qType.Field(i).Tag.Lookup(FromQueryTag)
		if !ok {
			//递归调用
			ResolveSearchQuery(driver, qValue.Field(i).Interface(), condition)
			continue
		}
		switch tag {
		case "-":
			continue
		}
		t = makeTag(tag)

		// 跳过空值
		if qValue.Field(i).IsZero() {
			continue
		}
		// 解析
		switch t.Type {
		case "left":
			// 左关联
			join := condition.SetJoinOn(t.Type, fmt.Sprintf(
				"left join `%s` on `%s`.`%s` = `%s`.`%s`",
				t.Join,
				t.Join,
				t.On[0],
				t.Table,
				t.On[1],
			))
			ResolveSearchQuery(driver, qValue.Field(i).Interface(), join)
		case "eq", "exact", "iexact":
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` = ?", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
		case "like", "contains", "icontains":
			// fixme mysql不支持ilike
			if driver == Postgres && (t.Type == "icontains" || t.Type == "like") {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` ilike ?", t.Table, t.Column), []interface{}{"%" + qValue.Field(i).String() + "%"})
			} else {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{"%" + qValue.Field(i).String() + "%"})
			}
		case "gt":
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` > ?", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
		case "gte":
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` >= ?", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
		case "lt":
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` < ?", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
		case "lte":
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` <= ?", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
		case "startswith", "istartswith":
			if driver == Postgres && t.Type == "istartswith" {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` ilike ?", t.Table, t.Column), []interface{}{qValue.Field(i).String() + "%"})
			} else {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{qValue.Field(i).String() + "%"})
			}
		case "endswith", "iendswith":
			if driver == Postgres && t.Type == "iendswith" {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` ilike ?", t.Table, t.Column), []interface{}{"%" + qValue.Field(i).String()})
			} else {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{"%" + qValue.Field(i).String()})
			}
		// between 介于两者之间
		case "between":
			// 判断是否为字符串切片类型 并且长度为2
			if qValue.Field(i).Kind() == reflect.Slice && len(qValue.Field(i).Interface().([]string)) == 2 {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` between ? and ?", t.Table, t.Column), []interface{}{
					qValue.Field(i).Index(0).String(),
					qValue.Field(i).Index(1).String(),
				})
			}
		case "in":
			// 判断值长度大于0
			if qValue.Field(i).Kind() == reflect.Slice && qValue.Field(i).Len() > 0 {
				condition.SetWhere(fmt.Sprintf("`%s`.`%s` in (?)", t.Table, t.Column), []interface{}{qValue.Field(i).Interface()})
			}
		case "isnull":
			if !(qValue.Field(i).IsZero() && qValue.Field(i).IsNil()) {
				strVal := qValue.Field(i).String()
				if strVal == "0" || strVal == "false" {
					condition.SetWhere(fmt.Sprintf("`%s`.`%s` is null", t.Table, t.Column), make([]interface{}, 0))
				} else {
					condition.SetWhere(fmt.Sprintf("`%s`.`%s` is not null", t.Table, t.Column), make([]interface{}, 0))
				}
			}
		case "order":
			switch strings.ToLower(qValue.Field(i).String()) {
			case "desc", "asc":
				condition.SetOrder(fmt.Sprintf("`%s`.`%s` %s", t.Table, t.Column, qValue.Field(i).String()))
			}
		case "page":
			condition.SetPage(fmt.Sprintf("%v", qValue.Field(i).Interface()))
		case "pageSize":
			condition.SetPageSize(fmt.Sprintf("%v", qValue.Field(i).Interface()))
		}
	}
}

var (
	Source string
	Driver string
	DBName string
)

// 生成搜索条件 ｜ 备注: 传入的参数不能为指针类型
// exact / iexact 等于
// contains / icontains 包含
// gt / gte 大于 / 大于等于
// lt / lte 小于 / 小于等于
// startswith / istartswith 以…起始
// endswith / iendswith 以…结束
// in
// isnull / 0或false >> is null  1或true >> is not null
// order 排序		e.g. order[key]=desc     order[key]=asc
func MakeCondition(q interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		condition := &GormCondition{
			GormPublic: GormPublic{},
			Join:       make([]*GormJoin, 0),
		}
		ResolveSearchQuery(Driver, q, condition)
		for _, join := range condition.Join {
			if join == nil {
				continue
			}
			db = db.Joins(join.JoinOn)
			for k, v := range join.Where {
				db = db.Where(k, v...)
			}
			for k, v := range join.Or {
				db = db.Or(k, v...)
			}
			for _, o := range join.Order {
				db = db.Order(o)
			}
		}
		for k, v := range condition.Where {
			db = db.Where(k, v...)
		}
		for k, v := range condition.Or {
			db = db.Or(k, v...)
		}
		for _, o := range condition.Order {
			db = db.Order(o)
		}
		if condition.Page != "" && condition.PageSize != "" {
			// 查询全部
			if condition.PageSize == "-1" {
				db = db.Offset(0).Limit(-1)
			} else {
				page, _ := strconv.Atoi(condition.Page)
				pageSize, _ := strconv.Atoi(condition.PageSize)
				if page <= 0 {
					page = 1
				}
				if pageSize <= 0 {
					pageSize = 10
				}

				offset := (page - 1) * pageSize
				if offset < 0 {
					offset = 0
				}
				db = db.Offset(int(offset)).Limit(int(pageSize))
			}
		}
		return db
	}
}

// 生成分页scope | 废弃，以融合到上面一个函数里
func Paginate(page, pageSize int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 查询全部
		if pageSize == -1 {
			return db.Offset(0).Limit(-1)
		}

		// 常规
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		if offset < 0 {
			offset = 0
		}
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
