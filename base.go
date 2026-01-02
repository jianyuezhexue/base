package base

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/jianyuezhexue/base/db"
	"github.com/jianyuezhexue/base/localCache"
	"github.com/jianyuezhexue/base/tool"
	"github.com/jinzhu/copier"
	"github.com/looplab/fsm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var OmitCreateFileds = []string{"created_at", "create_by", "create_by_name"}
var OmitUpdateFileds = []string{"updated_at", "update_by", "update_by_name"}

// 底层类型约定
type SearchCondition = func(db *gorm.DB) *gorm.DB
type PreloadsType = map[string][]any
type RecordLogFunc = func(ctx *gin.Context, operatorType, operatorTypeName string, oldData, newData any) error

var safeColumnNameRegex = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// 白名单校验列名，防止 filedName 被拼接进 SQL 时产生注入风险。
// 允许格式: column 或 table.column（多段也允许），仅允许字母/数字/下划线/点。
func validateSafeColumnName(column string) error {
	if column == "" {
		return fmt.Errorf("字段名不能为空")
	}
	if !safeColumnNameRegex.MatchString(column) {
		return fmt.Errorf("字段名非法: %s", column)
	}
	return nil
}

// 充血模型基础接口
type BaseModelInterface[T any] interface {
	TableName() string                                                                                                               // 表名
	Tx() *gorm.DB                                                                                                                    // 获取事务DB
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error                                                            // 事务处理
	SetData(data any) (*T, error)                                                                                                    // 设置数据
	Validate() error                                                                                                                 // 数据校验
	Complete() error                                                                                                                 // 完善数据
	Create() (*T, error)                                                                                                             // 新增数据
	CreateWithData(*T) (*T, error)                                                                                                   // 保存数据
	Update() (*T, error)                                                                                                             // 更新数据
	UpdateWithData(data *T) (*T, error)                                                                                              // 使用传入对象更新数据
	LoadData(cond SearchCondition, preloads ...PreloadsType) (*T, error)                                                             // 加载数据
	LoadById(id uint64, preloads ...PreloadsType) (*T, error)                                                                        // 根据Id加载数据
	LoadByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) (*T, error)                                           // 根据业务编码查询数据
	GetById(Id uint64, preloads ...PreloadsType) (*T, error)                                                                         // 根据Id查询数据
	GetByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error)                                                                   // 根据Id查询数据
	Repair() error                                                                                                                   // 修复数据
	Count(conds ...SearchCondition) (int64, error)                                                                                   // 统计数据条数
	List(conds ...SearchCondition) ([]*T, error)                                                                                     // 查询列表数据
	ListByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error)                                                                  // 根据Ids查询数据
	ListByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) ([]*T, error)                                         // 根据业务编码查询列表数据
	CountByBusinessCode(filedName, filedValue string) (int64, error)                                                                 // 根据业务编码统计数量
	ListByBusinessCodes(filedName string, filedValues []string, preloads ...PreloadsType) ([]*T, error)                              // 根据业务编码列表查询数据
	CountByBusinessCodes(filedName string, filedValues []string) (int64, error)                                                      // 根据业务编码列表统计数量
	MaxId() (int64, error)                                                                                                           // 获取最大ID
	Del(ids ...uint64) error                                                                                                         // 删除数据
	CheckBusinessCodeExist(filedName, businessCode string) (bool, error)                                                             // 检查业务编码是否重复
	BusinessCodeCannotRepeat(filedName, businessCode string) error                                                                   // 业务编码不能重复
	CheckBusinessCodesExist(filedName string, values []string) (map[int]bool, error)                                                 // 批量检查业务编码是否存在
	CheckUniqueKeysExist(filedNames []string, values []string) (bool, error)                                                         // 检查唯一键是否重复
	CheckUniqueKeysExistBatch(filedNames []string, values [][]string, withOutIds ...uint64) ([]bool, error)                          // 批量检查唯一键是否重复
	MakeConditon(data any) func(db *gorm.DB) *gorm.DB                                                                                // 构造查询条件
	ReInit(entity *T, baseModel *BaseModel[T]) error                                                                                 // 重置模型中的Context和Db
	InitStateMachine(initStatus string, events []fsm.EventDesc, afterEvent fsm.Callback, callbacks ...map[string]fsm.Callback) error // 初始化状态机
	EventExecution(initStatus, event, eventZhName string, args ...any) error                                                         // 执行事件
}

// 公共模型属性
type BaseModel[T any] struct {
	Id                    uint64            `json:"id" uri:"id" search:"-" gorm:"primarykey"` // 主键
	CreateBy              string            `json:"createBy" gorm:"<-:create" search:"-"`     // 创建人
	CreateByName          string            `json:"createByName" gorm:"<-:create" search:"-"` // 创建人名称
	CreatedAt             db.LocalTime      `json:"createdAt" gorm:"<-:create"  search:"-"`   // 创建时间
	UpdateBy              string            `json:"updateBy" gorm:"<-:update" search:"-"`     // 更新人
	UpdateByName          string            `json:"updateByName" gorm:"<-:update" search:"-"` // 更新人名称
	UpdatedAt             db.LocalTime      `json:"updatedAt" gorm:"<-:update" search:"-"`    // 更新时间
	DeletedAt             gorm.DeletedAt    `json:"-" gorm:"index" search:"-"`                // 删除标记
	Db                    *gorm.DB          `json:"-" gorm:"-" search:"-"`                    // 数据库连接
	Ctx                   *gin.Context      `json:"-" gorm:"-" search:"-"`                    // 上下文
	Preloads              map[string][]any  `json:"-" gorm:"-" search:"-"`                    // 预加载
	TableName             string            `json:"-" gorm:"-" search:"-"`                    // 表名
	OperatorId            string            `json:"-" gorm:"-" search:"-"`                    // 操作日志操作人id
	OperatorName          string            `json:"-" gorm:"-" search:"-"`                    // 操作日志操作人
	CustomerOrder         string            `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`  // 自定义排序规则
	DefaultSearchConditon SearchCondition   `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`  // 默认搜索条件
	PermissionConditons   []SearchCondition `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`  // 权限条件
	StatesMachine         *fsm.FSM          `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`  // 状态机
	EntityKey             string            `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`  // 业务实体Key
}

// 初始化模型
func NewBaseModel[T any](ctx *gin.Context, db *gorm.DB, tableName string, entity *T) BaseModel[T] {

	// 前置校验
	if ctx == nil {
		panic("调用[NewBaseModel]入参, ctx为nil,请开发检查")
	}
	if ctx.Request == nil {
		panic("调用[NewBaseModel]入参, ctx.Request is nil,请开发检查")
	}
	if entity == nil {
		panic("调用[NewBaseModel]入参, 传入的entity为nil,请开发检查")
	}

	// 从上下文中读取当前用户信息
	userId, _ := ctx.Get("currUserId")
	userName, _ := ctx.Get("currUserName")

	// 基础模型赋值
	entityKey := fmt.Sprintf("%p", entity) // 实体指针地址
	baseModel := BaseModel[T]{
		Ctx:       ctx,
		Db:        db,
		TableName: tableName,
		EntityKey: entityKey,
	}

	// 读取Context中的deadline
	remaining := 5 * time.Minute
	deadline, ok := ctx.Request.Context().Deadline()
	if ok {
		// 计算剩余时间
		remaining = time.Until(deadline)
	}

	// 将业务模型放到本地缓存中 | 过期时间: context中剩余时间或默认5分钟过期
	localCache := localCache.NewCache()
	localCache.Set(entityKey, entity, remaining)

	// 从Ctx中读取用户信息
	baseModel.OperatorId = fmt.Sprintf("%v", userId)
	baseModel.OperatorName = fmt.Sprintf("%v", userName)

	// 在db context 预埋用户信息
	dbContet := context.Background()
	dbContet = context.WithValue(dbContet, "currUserId", userId)
	dbContet = context.WithValue(dbContet, "currUserName", userName)
	// baseModel.Db.Statement.Context = dbContet
	baseModel.Db = baseModel.Db.WithContext(dbContet)

	// 给一个空默认搜索条件
	baseModel.DefaultSearchConditon = func(db *gorm.DB) *gorm.DB {
		return db
	}

	return baseModel
}

// ---------- OPTIONS函数 ----------
type Option[T any] func(*BaseModel[T])

// 初始化带上权限条件
func WithPermissionConditons[T any](conds ...SearchCondition) Option[T] {
	return func(b *BaseModel[T]) {
		b.PermissionConditons = conds
	}
}

// 注入Preload
func WithPreloads[T any](preloads map[string][]any) Option[T] {
	return func(b *BaseModel[T]) {
		b.Preloads = preloads
	}
}

// 自定义排序规则
func WithCustomerOrder[T any](order string) Option[T] {
	return func(b *BaseModel[T]) {
		b.CustomerOrder = order
	}
}

// ---------- 公共底层业务函数 ----------

// 记录操作日志
const LogTypeCreate string = "create"
const LogTypeUpdate string = "update"
const LogTypeDelete string = "delete"

// 记录操作日志 ｜ todo
func (b *BaseModel[T]) RecordLog(operatorType, operatorTypeName string, oldData, newData any) error {
	// todo

	return nil
}

// 数据校验钩子函数
func (b *BaseModel[T]) Validate() error {
	return nil
}

// 数据修复
func (b *BaseModel[T]) Repair() error {
	return nil
}

// 数据完善
func (b *BaseModel[T]) Complete() error {
	return nil
}

// 获取当前时间
func (b *BaseModel[T]) CurrTime() time.Time {
	var currTime time.Time
	// 从Ctx中读取当前时间
	ctxCurrTime, _ := b.Ctx.Get("CurrTime")
	if ctxCurrTime != nil {
		return ctxCurrTime.(time.Time)
	}

	// 如果没有手动设置
	currTime = time.Now().Local() // 当前时间
	b.Ctx.Set("CurrTime", currTime)
	return currTime
}

// 获取当前业务实体
func (b *BaseModel[T]) GetCurrEntity() (*T, error) {
	// 从本地缓存中读取
	localCache := localCache.NewCache()
	entity, exist := localCache.Get(b.EntityKey)
	if !exist {
		return nil, fmt.Errorf("本地缓存中没有[%v]对应的业务实体,请开发检查", b.EntityKey)
	}

	// 断言判断
	resEntity, ok := entity.(*T)
	if !ok {
		return nil, fmt.Errorf("本地缓存中没有[%v]对应的业务实体断言失败，请检查", b.EntityKey)
	}

	return resEntity, nil
}

// 构造查询条件 | 这里不能传指针注意
func (b *BaseModel[T]) MakeConditon(data any) SearchCondition {
	return db.MakeCondition(data)
}

// 清空搜索条件
// 清除分页和偏移量
func (b *BaseModel[T]) ClearOffset() SearchCondition {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Limit(-1).Offset(-1)
		return db
	}
}

// 按照Id倒叙查询
func (b *BaseModel[T]) OrderByIdDesc() SearchCondition {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Order("id desc")
		return db
	}
}

// 按照Id生序查询
func (b *BaseModel[T]) OrderByIdAsc() SearchCondition {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Order("id asc")
		return db
	}
}

// 设置数据
func (b *BaseModel[T]) SetData(data any) (*T, error) {
	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 初始化实体对象
	err = tool.CopyDeep(entity, data)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// 创建数据 | 将自身作为存储对象
func (b *BaseModel[T]) Create() (*T, error) {

	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 执行创建操作
	err = b.Tx().Omit(OmitUpdateFileds...).Create(entity).Error
	if err != nil {
		return nil, err
	}

	// 记录日志
	err = b.RecordLog(LogTypeCreate, "新增", new(T), entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// 创建数据 | 使用传入对象作为存储对象
func (b *BaseModel[T]) CreateWithData(data *T) (*T, error) {
	// 执行创建操作
	err := b.Tx().Omit(OmitUpdateFileds...).Create(data).Error
	if err != nil {
		return nil, err
	}
	// 记录日志
	err = b.RecordLog(LogTypeCreate, "新增", new(T), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 更新数据 | 将自身作为更新对象
func (b *BaseModel[T]) Update() (*T, error) {
	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 执行更新操作
	session := &gorm.Session{FullSaveAssociations: true, Context: b.Db.Statement.Context}
	err = b.Tx().Omit(OmitCreateFileds...).Session(session).Clauses(clause.OnConflict{UpdateAll: true}).Save(entity).Error
	if err != nil {
		return nil, err
	}

	// 记录日志
	// TODO 这里没有区分新旧数据，后续需要优化
	err = b.RecordLog(LogTypeUpdate, "更新", entity, entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// 更新数据 | 使用传入对象作为更新对象
func (b *BaseModel[T]) UpdateWithData(data *T) (*T, error) {
	// 执行更新操作
	session := &gorm.Session{FullSaveAssociations: true, Context: b.Db.Statement.Context}
	err := b.Tx().Omit(OmitCreateFileds...).Session(session).Clauses(clause.OnConflict{UpdateAll: true}).Save(data).Error
	if err != nil {
		return nil, err
	}

	// 记录日志
	// TODO 这里没有区分新旧数据，后续需要优化
	err = b.RecordLog(LogTypeUpdate, "更新", data, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 删除数据
func (b *BaseModel[T]) Del(ids ...uint64) error {
	// 执行删除操作
	model := new(T)
	err := b.Tx().Where("id in ?", ids).Delete(model).Error
	if err != nil {
		return err
	}
	return nil
}

// 统计数据条数 | 搜索条件: 默认条件,权限条件,搜索条件,拓展搜索条件
func (b *BaseModel[T]) Count(conds ...SearchCondition) (int64, error) {
	var total int64
	err := b.Db.Debug().Model(new(T)).
		Scopes(b.DefaultSearchConditon).
		Scopes(b.PermissionConditons...).
		Scopes(conds...).
		Scopes(b.ClearOffset()).
		Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, err
}

// 查询列表数据 | 搜索条件: 默认条件,权限条件,搜索条件,拓展搜索条件
func (b *BaseModel[T]) List(conds ...SearchCondition) ([]*T, error) {

	// 组合查询条件
	db := b.Db.Debug().
		Scopes(b.DefaultSearchConditon).  // 默认条件
		Scopes(b.PermissionConditons...). // 权限条件
		Scopes(conds...)                  // 搜索条件

	// 自定义排序规则
	if b.CustomerOrder != "" {
		db = db.Order(b.CustomerOrder) // 自定义排序
	} else {
		db = db.Order("id desc") // 默认排序
	}

	// 预加载查询
	if len(b.Preloads) > 0 {
		for key, vals := range b.Preloads {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id desc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 执行查询
	var list []*T
	err := db.Find(&list).Error
	if err != nil {
		return nil, err
	}

	return list, err
}

// 加载数据
func (b *BaseModel[T]) LoadData(cond SearchCondition, preloads ...PreloadsType) (*T, error) {

	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 预加载查询
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	err = db.Scopes(cond).First(entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("[%s]查询的数据不存在,请检查", b.TableName)
		}
		return nil, err
	}

	return entity, nil
}

// 根据Id加载数据
func (b *BaseModel[T]) LoadById(id uint64, preloads ...PreloadsType) (*T, error) {

	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 预加载查询
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	err = db.Where("id = ?", id).First(entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("[%v]查询的数据不存在,请检查", b.TableName)
		}
		return entity, err
	}

	return entity, nil
}

// 根据业务单号查询数据
func (b *BaseModel[T]) LoadByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) (*T, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return nil, err
	}
	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return nil, fmt.Errorf("[BASE]中业务实体为空,请开发检查")
	}

	// 预加载查询
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	err = db.Where(fmt.Sprintf("%s = ?", filedName), filedValue).First(entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("[%v]对应业务Code[%s:%s]查询的数据不存在,请检查", b.TableName, filedName, filedValue)
		}
		return entity, err
	}
	return entity, nil
}

// 根据Id查询数据
func (b *BaseModel[T]) GetById(Id uint64, preloads ...PreloadsType) (*T, error) {
	// 预加载查询
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	data := new(T)
	err := db.Where("id = ?", Id).First(data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询的数据不存在,请检查")
		}
		return nil, err
	}
	return data, nil
}

// 根据Ids查询数据
func (b *BaseModel[T]) GetByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error) {

	// 预加载处理
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 组合查询条件
	db = db.Where("id in ?", Ids)

	// 组合排序规则
	if b.CustomerOrder != "" {
		db = db.Order(b.CustomerOrder)
	} else {
		db = db.Order("id asc")
	}

	// 数据查询
	dataList := []*T{}
	err := db.Debug().Find(&dataList).Error
	if err != nil {
		return nil, err
	}
	return dataList, nil
}

// 根据Ids查询数据
func (b *BaseModel[T]) ListByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error) {
	// 预加载处理
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	dataList := make([]*T, 0)
	err := db.Where("id in ?", Ids).Find(&dataList).Error
	if err != nil {
		return nil, err
	}
	return dataList, nil
}

// 根据业务编码查询列表
func (b *BaseModel[T]) ListByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) ([]*T, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return nil, err
	}
	// 预加载处理
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	list := []*T{}
	err := db.Where(fmt.Sprintf("%s = ?", filedName), filedValue).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// 根据业务编码列表查询列表
func (b *BaseModel[T]) ListByBusinessCodes(filedName string, filedValues []string, preloads ...PreloadsType) ([]*T, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return nil, err
	}
	if len(filedValues) == 0 {
		return nil, fmt.Errorf("ListByBusinessCodes查询,业务编码列表不能为空")
	}

	// 预加载处理
	db := b.Db
	if len(preloads) > 0 {
		for key, vals := range preloads[0] {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// filedValues 为空时，避免生成 in () 的无效 SQL
	list := make([]*T, 0)
	err := db.Where(fmt.Sprintf("%s in ?", filedName), filedValues).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CountByBusinessCode 根据业务编码统计数量
func (m *BaseModel[T]) CountByBusinessCode(filedName, filedValue string) (int64, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return 0, err
	}
	var count int64
	err := m.Db.Model(new(T)).Where(fmt.Sprintf("%s = ?", filedName), filedValue).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountByBusinessCodes 根据业务编码列表统计数量
func (m *BaseModel[T]) CountByBusinessCodes(filedName string, filedValues []string) (int64, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return 0, err
	}
	if len(filedValues) == 0 {
		return 0, fmt.Errorf("CountByBusinessCodes查询,业务编码列表不能为空")
	}
	var count int64
	err := m.Db.Model(new(T)).Where(fmt.Sprintf("%s in ?", filedName), filedValues).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MaxId 获取最大ID
func (m *BaseModel[T]) MaxId() (int64, error) {
	var maxId int64
	err := m.Db.Model(new(T)).Select("max(id)").Scan(&maxId).Error
	if err != nil {
		return 0, err
	}
	return maxId, nil
}

// 重置上下文和Db
func (b *BaseModel[T]) ReInit(entity *T, baseModel *BaseModel[T]) error {
	if b.Ctx == nil || b.Db == nil {
		return fmt.Errorf("[ReInit]Context或DB为空,请开发检查")
	}
	baseModel.Ctx = b.Ctx
	baseModel.Db = b.Db
	baseModel.TableName = b.TableName
	baseModel.EntityKey = fmt.Sprintf("%p", entity) // 实体指针地址
	return nil
}

//	校验业务单号是否存在
//
// 如果当前业务实体Id存在(意味着当前数据已经落库,会跳过当前)
// true 存在 false 不存在
func (b *BaseModel[T]) CheckBusinessCodeExist(filedName, businessCode string) (bool, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return true, err
	}
	ids := []uint64{}
	err := b.Db.Model(new(T)).Select("id").Where(fmt.Sprintf("%s = ?", filedName), businessCode).Find(&ids).Error
	if err != nil {
		return true, err
	}
	// 长度为0 绝对不存在
	if len(ids) == 0 {
		return false, nil
	}
	// 大于等于2 一定存在
	if len(ids) >= 2 {
		return true, nil
	}
	// 长度为1,如果是当前数据,则不存在
	if len(ids) == 1 && b.Id == ids[0] {
		return false, nil
	} else {
		return true, nil
	}
}

// 业务单号不可以重复
func (b *BaseModel[T]) BusinessCodeCannotRepeat(filedName, businessCode string) error {
	exist, err := b.CheckBusinessCodeExist(filedName, businessCode)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("[%v]业务单号[%s]不可重复,请检查", b.TableName, businessCode)
	}
	return nil
}

// 批量校验业务数据是否存在
func (b *BaseModel[T]) CheckBusinessCodesExist(filedName string, values []string) (map[int]bool, error) {
	if err := validateSafeColumnName(filedName); err != nil {
		return nil, err
	}
	res := make(map[int]bool)

	// 查询DB数据
	dbFileds := []string{}
	model := new(T)
	err := b.Db.Model(model).Select(filedName).Where(fmt.Sprintf("%s in ?", filedName), values).Find(&dbFileds).Error
	if err != nil {
		return res, err
	}

	// 对比数据并标记结果
	dbMap := make(map[string]struct{})
	for _, val := range dbFileds {
		dbMap[val] = struct{}{}
	}
	for i, v := range values {
		res[i] = false
		if _, exists := dbMap[v]; exists {
			res[i] = true
		}
	}
	return res, nil
}

//	校验唯一键是否存在 | 单条校验
//
// 如果当前业务实体Id存在(意味着当前数据已经落库,会跳过当前)
// true 存在 false 不存在
func (b *BaseModel[T]) CheckUniqueKeysExist(filedNames []string, values []string) (bool, error) {
	ids := []uint64{}
	stringBuilder := fmt.Sprintf("(%v) = ?", strings.Join(filedNames, ","))
	err := b.Db.Model(new(T)).Where(stringBuilder, values).Find(&ids).Error
	if err != nil {
		return true, err
	}
	// 长度为0 绝对不存在
	if len(ids) == 0 {
		return false, nil
	}
	// 大于等于2 一定存在
	if len(ids) >= 2 {
		return true, nil
	}
	// 长度为1,如果是当前数据,则不存在
	if len(ids) == 1 && b.Id == ids[0] {
		return false, nil
	} else {
		return true, nil
	}
}

//	批量校验唯一键是否存在 | 多条校验
//
// CONCAT_WS(",",order_id,status,create_by) as UniqueValues
// true 存在 false 不存在
func (b *BaseModel[T]) CheckUniqueKeysExistBatch(filedNames []string, values [][]string, withOutIds ...uint64) ([]bool, error) {
	res := make([]bool, len(values))
	if len(values) == 0 || len(filedNames) == 0 {
		return res, nil
	}

	// 定义结构体
	type itemData struct {
		Id           uint64 // 主键ID
		UniqueValues string // 逗号隔开的字符串拼接 ｜ CONCAT_WS
	}

	// 构建查询条件
	fieldsWithNull := make([]string, len(filedNames))
	for i, f := range filedNames {
		fieldsWithNull[i] = fmt.Sprintf("IFNULL(%s, '')", f)
	}
	whereBuilder := fmt.Sprintf("(%v) in ?", strings.Join(filedNames, ","))
	selectBuilder := fmt.Sprintf("id,CONCAT_WS(',',%v) as UniqueValues", strings.Join(fieldsWithNull, ","))

	// 执行查询
	list := []*itemData{}
	err := b.Db.Model(new(T)).Select(selectBuilder).Where(whereBuilder, values).Find(&list).Error
	if err != nil {
		return res, err
	}

	// 构建结果数据Map
	resMap := make(map[string]uint64)
	for _, item := range list {
		resMap[item.UniqueValues] = item.Id
	}

	// 对比数据处理
	for index, itemVals := range values {
		itemUniqueVal := strings.Join(itemVals, ",")

		// 结果中查询是否存在
		id, exists := resMap[itemUniqueVal]

		// 如果存在,且没有在withOutIds中,则认为重复
		if exists {
			inWithOutIds := slices.Contains(withOutIds, id)
			// 如果没有排除的ID,则认为重复
			if !inWithOutIds {
				res[index] = true
				continue
			}
		}

		// 默认不存在
		res[index] = false
	}

	return res, nil
}

// ---------- 事件驱动相关 ----------

// 初始化状态机
func (b *BaseModel[T]) InitStateMachine(initStatus string, events []fsm.EventDesc, afterEvent fsm.Callback, callbacks ...map[string]fsm.Callback) error {
	finelCallbacks := make(map[string]fsm.Callback)
	finelCallbacks["after_event"] = afterEvent
	if len(callbacks) > 0 {
		for _, item := range callbacks {
			for k, v := range item {
				finelCallbacks[k] = v
			}
		}
	}
	b.StatesMachine = fsm.NewFSM(initStatus, events, finelCallbacks)
	return nil
}

// 事件执行
func (b *BaseModel[T]) EventExecution(initStatus, event, eventZhName string, args ...any) error {
	// 0. 前置校验
	if b.StatesMachine == nil {
		return fmt.Errorf("状态机未注册,请开发检查")
	}

	// 读取业务实体 | 校验是否为空
	entity, err := b.GetCurrEntity()
	if err != nil {
		return fmt.Errorf("业务实体为空,请开发检查")
	}

	// 1. 重新设置初始状态
	b.StatesMachine.SetState(initStatus)

	// 2. 校验是否允许执行当前事件
	if !b.StatesMachine.Can(event) {
		return fmt.Errorf("业务实体[%s]当前状态[%s],不允许执行事件[%s],请开发检查", b.TableName, initStatus, eventZhName)
	}

	// 记录旧数据
	oldData := new(T)
	copier.Copy(oldData, entity)

	// 执行事件 | 注意状态没有变化是允许的
	ctx := b.Ctx.Request.Context()
	err = b.StatesMachine.Event(ctx, event, args)
	noTransitionError := fsm.NoTransitionError{Err: nil}
	if err != nil && !errors.Is(err, noTransitionError) {
		return fmt.Errorf("业务实体[%s]执行事件[%s]失败[%s],请开发检查", b.TableName, eventZhName, err.Error())
	}

	// 保存最新状态
	err = b.Tx().Save(entity).Error
	if err != nil {
		return fmt.Errorf("业务实体[%s]保存最终状态失败,请开发检查", b.TableName)
	}

	// 记录操作日志
	b.RecordLog(event, eventZhName, oldData, entity)
	return nil
}

// ---------- 事务函数 ----------

// 获取事务Db
func (m *BaseModel[T]) Tx() *gorm.DB {
	db, exist := m.Ctx.Get("txDb")
	if exist && db != nil {
		return db.(*gorm.DB)
	}
	return m.Db
}

// 开启事务
func (m *BaseModel[T]) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {

	// 防止重复开启事务
	_, exist := m.Ctx.Get("txDb")
	if exist {
		return fmt.Errorf("事务已开启,不要重复开启事务,请开发检查")
	}

	// 开启事务
	err := m.Db.Transaction(func(tx *gorm.DB) error {
		// 预埋事务Db
		m.Ctx.Set("txDb", tx)

		// 执行事务逻辑代码
		if err := fc(tx); err != nil {
			return err
		}

		// 回收事务Db
		m.Ctx.Set("txDb", nil)
		return nil
	}, opts...)
	return err
}

// 检查是否已经开启事务
func (m *BaseModel[T]) IsInTransaction() bool {
	_, exist := m.Ctx.Get("txDb")
	return exist
}

// ---------- 底层钩子 ----------

// 创建前钩子函数
func (b *BaseModel[T]) BeforeCreate(tx *gorm.DB) (err error) {

	ctx := tx.Statement.Context

	// 信息读取
	currUserId := ctx.Value("currUserId")
	if currUserId == nil || currUserId == "" {
		return fmt.Errorf("Ctx中[currUserId]不存在,请开发检查")
	}
	currUserName := ctx.Value("currUserName")
	if currUserName == nil || currUserName == "" {
		return fmt.Errorf("Ctx中[currUserName]不存在,请开发检查")
	}

	// 自动维护创建人信息
	if b.Id == 0 {
		b.CreateBy = currUserId.(string)
		b.CreateByName = currUserName.(string)
	} else {
		b.UpdateBy = currUserId.(string)
		b.UpdateByName = currUserName.(string)
	}
	b.OperatorId = currUserId.(string)
	b.OperatorName = currUserName.(string)
	return nil
}

// 更新前钩子函数
func (b *BaseModel[T]) BeforeUpdate(tx *gorm.DB) (err error) {

	ctx := tx.Statement.Context

	// 信息读取
	currUserId := ctx.Value("currUserId")
	if currUserId == nil || currUserId == "" {
		return fmt.Errorf("Ctx中[currUserId]不存在,请开发检查")
	}
	currUserName := ctx.Value("currUserName")
	if currUserName == nil || currUserName == "" {
		return fmt.Errorf("Ctx中[currUserName]不存在,请开发检查")
	}

	// 自动维护创建人信息
	b.UpdateBy = currUserId.(string)
	b.UpdateByName = currUserName.(string)
	b.OperatorId = currUserId.(string)
	b.OperatorName = currUserName.(string)
	return nil
}

// Save前钩子函数
func (b *BaseModel[T]) BeforeSave(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	// 信息读取
	currUserId := ctx.Value("currUserId")
	if currUserId == nil || currUserId == "" {
		return fmt.Errorf("Ctx中[currUserId]不存在,请开发检查")
	}
	currUserName := ctx.Value("currUserName")
	if currUserName == nil || currUserName == "" {
		return fmt.Errorf("Ctx中[currUserName]不存在,请开发检查")
	}

	// 自动维护创建人信息
	if b.Id == 0 {
		// 新建
		b.CreateBy = currUserId.(string)
		b.CreateByName = currUserName.(string)
	}
	if b.Id != 0 {
		// 更新
		b.UpdateBy = currUserId.(string)
		b.UpdateByName = currUserName.(string)
	}
	b.OperatorId = currUserId.(string)
	b.OperatorName = currUserName.(string)

	return nil
}

// 删除前钩子函数
func (b *BaseModel[T]) BeforeDelete(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context

	// 信息读取
	currUserId := ctx.Value("currUserId")
	if currUserId == nil || currUserId == "" {
		return fmt.Errorf("Ctx中[currUserId]不存在,请开发检查")
	}
	currUserName := ctx.Value("currUserName")
	if currUserName == nil || currUserName == "" {
		return fmt.Errorf("Ctx中[currUserName]不存在,请开发检查")
	}

	// 自动维护创建人信息
	b.UpdateBy = currUserId.(string)
	b.UpdateByName = currUserName.(string)
	b.OperatorId = currUserId.(string)
	b.OperatorName = currUserName.(string)

	return nil
}

// ---------- Ctx缓存 ----------

// 设置缓存，增加防并发锁
func GetDataWithCtxCache[T any](ctx *gin.Context, key string, fn func() (T, error)) (T, error) {

	// 使用互斥锁防止并发
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	// 先判断Ctx中是否有数据
	if data, ok := ctx.Get(key); ok {
		return data.(T), nil
	}

	// 执行函数
	data, err := fn()
	if err != nil {
		var zero T
		return zero, err
	}

	// 设置缓存
	ctx.Set(key, data)

	return data, nil
}

// 更新缓存
func ResetDataWithCtxCache[T any](ctx *gin.Context, key string, data T) {
	// 使用互斥锁防止并发
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	// 设置缓存
	ctx.Set(key, data)
}
