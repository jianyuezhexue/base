package base

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyuezhexue/base/db"
	"github.com/looplab/fsm"
	"gorm.io/gorm"
)

var OmitCreateFileds = []string{"created_at", "create_by", "create_by_name"}

// 底层类型约定
type SearchConditon = func(db *gorm.DB) *gorm.DB
type PreloadsType = map[string][]any
type RecordLogFunc = func(ctx *gin.Context, operatorType, operatorTypeName string, oldData, newData any) error

// 充血模型基础接口
type BaseModelInterface[T any] interface {
	TableName() string                                                                                       // 表名
	Tx() *gorm.DB                                                                                            // 获取事务DB
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error                                    // 事务处理
	SetData(data any) (*T, error)                                                                            // 设置数据
	Create() (*T, error)                                                                                     // 新增数据
	Update() (*T, error)                                                                                     // 更新数据
	Del(ids ...uint64) error                                                                                 // 删除数据
	LoadData(cond SearchConditon, preloads ...PreloadsType) (*T, error)                                      // 根据搜索条件加载数据
	LoadById(id uint64) (*T, error)                                                                          // 根据Id加载数据
	LoadByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) (*T, error)                   // 根据业务编码查询数据
	GetById(Id uint64, preloads ...PreloadsType) (*T, error)                                                 // 根据Id查询数据
	GetByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error)                                           // 根据Id查询数据
	Repair() error                                                                                           // 修复数据
	Count(SearchConditon, ...SearchConditon) (int64, error)                                                  // 统计数据条数
	List(SearchConditon, ...SearchConditon) ([]*T, error)                                                    // 查询列表数据
	Import(data any) error                                                                                   // 导入数据
	Export(cond SearchConditon) (string, error)                                                              // 导出数据
	Complete() error                                                                                         // 完善数据
	CheckBusinessCodeRepeat(filedName, businessCode string) (bool, error)                                    // 检查业务编码是否重复
	CheckBusinessCodesExist(filedName string, values []string, more ...SearchConditon) (map[int]bool, error) // 批量检查业务编码是否存在
	CheckUniqueKeysRepeat(filedNames []string, values []string, withOutIds ...uint64) (bool, error)          // 检查唯一键是否重复
	CheckUniqueKeysRepeatBatch(filedNames []string, values [][]string, withOutIds ...uint64) ([]bool, error) // 批量检查唯一键是否重复
	EventExecution(initStatus, event, eventZhName string, callbacks ...EventCallback[T]) error               // 执行事件 | 注意状态机一定要配置after_event回调
	ReInit(baseModel *BaseModel[T]) error                                                                    // 重新初始化
}

// 公共模型属性
type BaseModel[T any] struct {
	Id                    uint64           `json:"id" uri:"id" search:"exact" gorm:"primarykey"` // 主键
	CreatedAt             db.LocalTime     `json:"createdAt" search:"gte"`                       // 创建时间
	UpdatedAt             db.LocalTime     `json:"updatedAt" search:"lte"`                       // 更新时间
	DeletedAt             gorm.DeletedAt   `json:"-" gorm:"index" search:"-"`                    // 删除标记
	CreateBy              string           `json:"createBy" search:"eq"`                         // 创建人
	UpdateBy              string           `json:"updateBy" search:"eq"`                         // 更新人
	CreateByName          string           `json:"createByName" search:"eq"`                     // 创建人名称
	UpdateByName          string           `json:"updateByName" search:"eq"`                     // 更新人名称
	Db                    *gorm.DB         `json:"-" gorm:"-" search:"-"`                        // 数据库连接
	Ctx                   *gin.Context     `json:"-" gorm:"-" search:"-"`                        // 上下文
	Preloads              map[string][]any `json:"-" gorm:"-" search:"-"`                        // 预加载
	CurrTime              time.Time        `json:"-" gorm:"-" search:"-"`                        // 当前时间
	TableName             string           `json:"-" gorm:"-" search:"-"`                        // 表名
	OperatorId            string           `json:"-" gorm:"-" search:"-"`                        // 操作日志操作人id
	OperatorName          string           `json:"-" gorm:"-" search:"-"`                        // 操作日志操作人
	CustomerOrder         string           `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 自定义排序规则
	ParamData             any              `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 参数数据
	PermissionConditons   []SearchConditon `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 权限条件
	Entity                *T               `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 实体对象
	StatesMachine         *fsm.FSM         `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 状态机
	DefaultSearchConditon SearchConditon   `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 默认搜索条件
	RecordLogFunc         RecordLogFunc    `json:"-" gorm:"-" search:"-" copier:"-" vd:"-"`      // 记录日志函数
}

// ---------- OPTIONS函数 ----------
type Option[T any] func(*BaseModel[T])

// 初始化带上权限条件
func WithPermissionConditons[T any](conds ...SearchConditon) Option[T] {
	return func(b *BaseModel[T]) {
		b.PermissionConditons = conds
	}
}

// WithPreloads 注入Preload
func WithPreloads[T any](preloads map[string][]any) Option[T] {
	return func(b *BaseModel[T]) {
		b.Preloads = preloads
	}
}

// WithCustomerOrder 自定义排序规则
func WithCustomerOrder[T any](order string) Option[T] {
	return func(b *BaseModel[T]) {
		b.CustomerOrder = order
	}
}

// ---------- 公共底层业务函数 ----------

// 记录操作日志
const LogTypeCreate string = "create"
const LogTypeUpdate string = "update"

func (b *BaseModel[T]) RecordLog(operatorType, operatorTypeName string, oldData, newData any) error {
	if b.RecordLogFunc == nil {
		return errors.New("记录日志函数未初始化")
	}
	err := b.RecordLogFunc(b.Ctx, operatorType, operatorTypeName, oldData, newData)
	if err != nil {
		return err
	}
	return nil
}

// 构造查询条件 | 这里不能传指针注意 | 废弃中，直接使用db.MakeCondition()
func (b *BaseModel[T]) MakeConditon(data any) SearchConditon {
	return db.MakeCondition(data)
}

// 清空搜索条件
// 清除分页和偏移量
func (b *BaseModel[T]) ClearOffset() SearchConditon {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Limit(-1).Offset(-1)
		return db
	}
}

// 根据Id加载数据 LoadById(id uint64) (*T, error)
func (b *BaseModel[T]) LoadById(id uint64) (*T, error) {
	// 前置校验
	if b.Entity == nil {
		return nil, errors.New("初始化实体时候没有传进来,请开发检查")
	}

	// 预加载查询
	db := b.Db
	if len(b.Preloads) > 0 {
		for key, vals := range b.Preloads {
			// 组合where条件和order条件
			vals = append(vals, func(db *gorm.DB) *gorm.DB {
				return db.Order("id asc")
			})
			db = db.Preload(key, vals...)
		}
	}

	// 查询数据
	err := db.Where("id = ?", id).First(b.Entity).Error
	if err != nil {
		return b.Entity, err
	}

	return b.Entity, nil
}

// LoadByBusinessCode 根据业务单号查询数据
func (b *BaseModel[T]) LoadByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) (*T, error) {
	// 前置校验
	if b.Entity == nil {
		return nil, errors.New("业务实体,没有注册进来,请开发检查")
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
	err := db.Where(fmt.Sprintf("%s = ?", filedName), filedValue).First(b.Entity).Error
	if err != nil {
		return b.Entity, err
	}
	return b.Entity, nil
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

// GetByIds 根据Ids查询数据
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

// ReInit 重置上下文和Db
func (b *BaseModel[T]) ReInit(baseModel *BaseModel[T]) error {
	if b.Ctx == nil || b.Db == nil {
		return fmt.Errorf("[ReInit]Context或DB为空,请开发检查")
	}

	baseModel.Ctx = b.Ctx
	baseModel.Db = b.Db
	baseModel.TableName = b.TableName
	return nil
}

// 校验业务单号是否重复
func (b *BaseModel[T]) CheckBusinessCodeRepeat(filedName, businessCode string) (bool, error) {
	var count int64
	model := new(T)
	err := b.Db.Model(model).Where(fmt.Sprintf("%s = ?", filedName), businessCode).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, fmt.Errorf("业务单号[%v]重复,请检查", businessCode)
	}
	return true, nil
}

// 批量校验业务数据是否存在
func (b *BaseModel[T]) CheckBusinessCodesExist(filedName string, values []string, more ...SearchConditon) (map[int]bool, error) {
	res := make(map[int]bool)

	// 控制数量
	if len(values) > 500 {
		return nil, fmt.Errorf("批量校验业务数据是否存在,单次数量不能超过500个,请开发检查")
	}

	// 查询DB数据
	dbFileds := []string{}
	model := new(T)
	err := b.Db.Model(model).Select(filedName).Scopes(more...).Where(fmt.Sprintf("%s in ?", filedName), values).Find(&dbFileds).Error
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

// 校验唯一键是否存在 | 单条校验
func (b *BaseModel[T]) CheckUniqueKeysRepeat(filedNames []string, values []string, withOutIds ...uint64) (bool, error) {
	var count int64
	model := new(T)
	stringBuilder := fmt.Sprintf("(%v) = ?", strings.Join(filedNames, ","))
	// 排除自身 | todo 待这里做优化 not in 不能命中索引
	scopes := func(db *gorm.DB) *gorm.DB {
		if len(withOutIds) > 0 {
			return db.Where("id not in ?", withOutIds)
		}
		return db
	}
	err := b.Db.Model(model).Scopes(scopes).Where(stringBuilder, values).Count(&count).Error
	if err != nil {
		return true, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

// 批量校验唯一键是否存在 | 多条校验
func (b *BaseModel[T]) CheckUniqueKeysRepeatBatch(filedNames []string, values [][]string, withOutIds ...uint64) ([]bool, error) {
	// 控制数量
	if len(values) > 500 {
		return nil, fmt.Errorf("批量校验唯一键是否存在,单次数量不能超过500个,请开发检查")
	}

	// todo 这里有性能问题,待优化成批量查询,内存中做对比处理
	res := make([]bool, len(values))
	for i, v := range values {
		repeat, err := b.CheckUniqueKeysRepeat(filedNames, v, withOutIds...)
		if err != nil {
			return res, err
		}
		res[i] = repeat
	}
	return res, nil
}

// 执行某个事件
type EventCallback[T any] func() error

func (b *BaseModel[T]) EventExecution(initStatus, event, eventZhName string, callbacks ...EventCallback[T]) error {
	// 1.校验状态机是否已注册
	if b.StatesMachine == nil {
		return fmt.Errorf("状态机未注册,请开发检查")
	}

	// 2. todo 检查有没有配置  after_event 回调

	// 3. 重新设置初始状态
	b.StatesMachine.SetState(initStatus)

	// 4. 判断目标状态和当前状态是否一致,跳过

	// 校验是否允许执行当前事件
	if !b.StatesMachine.Can(event) {
		return fmt.Errorf("业务实体[%s]当前状态[%s],不允许执行事件[%s],请开发检查", b.TableName, initStatus, eventZhName)
	}

	if b.Entity == nil {
		return fmt.Errorf("BaseModel中业务实体为空,需要在实例化候传入,请开发检查")
	}

	// 记录旧数据
	oldData := b.Entity

	// 执行前置回调[0]
	if len(callbacks) >= 1 {
		err := callbacks[0]()
		if err != nil {
			return fmt.Errorf("业务实体[%s]执行事件[%s]前置回调失败,请开发检查", b.TableName, eventZhName)
		}
	}

	// 执行事件 && 触发钩子函数对实体进行状态修改 | 注意状态没有变化是允许的
	ctx := b.Ctx.Request.Context()
	err := b.StatesMachine.Event(ctx, event)
	noTransitionError := fsm.NoTransitionError{Err: nil}
	if err != nil && !errors.Is(err, noTransitionError) {
		return fmt.Errorf("业务实体[%s]执行事件[%s]失败[%s],请开发检查", b.TableName, eventZhName, err.Error())
	}

	// 执行后置回调[1]
	if len(callbacks) >= 2 {
		err := callbacks[1]()
		if err != nil {
			return fmt.Errorf("业务实体[%s]执行事件[%s]后置回调失败,请开发检查", b.TableName, eventZhName)
		}
	}

	// 保存状态
	err = b.Tx().Save(b.Entity).Error
	if err != nil {
		return fmt.Errorf("业务实体[%s]保存最终状态失败,请开发检查", b.TableName)
	}

	// 记录操作日志
	b.RecordLog(event, eventZhName, oldData, b.Entity)
	return nil
}

// ---------- 底层函数 ----------

// 获取事务Db | 使用Tx()代替DB()方法,语义更清晰
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
	if b.Ctx != nil {
		// 自动维护创建人信息
		userInfo, exise := b.Ctx.Get("identity")
		if exise {
			userInfo := userInfo.(map[string]any)
			b.CreateBy = userInfo["UserName"].(string)
			b.CreateByName = userInfo["NickName"].(string)
			b.OperatorId = userInfo["UserName"].(string)
			b.OperatorName = userInfo["NickName"].(string)
			fmt.Println("用户信息", userInfo)
		}
	}

	return nil
}

// 创建后钩子函数
func (b *BaseModel[T]) AfterCreate(tx *gorm.DB) (err error) {
	return nil
}

// 更新前钩子函数
func (b *BaseModel[T]) BeforeUpdate(tx *gorm.DB) (err error) {
	if b.Ctx != nil {
		// 自动维护更新人信息
		userInfo, exise := b.Ctx.Get("identity")
		if exise {
			userInfo := userInfo.(map[string]any)
			b.UpdateBy = userInfo["UserName"].(string)
			b.UpdateByName = userInfo["NickName"].(string)
			b.OperatorId = userInfo["UserName"].(string)
			b.OperatorName = userInfo["NickName"].(string)
		}
	}

	return nil
}

// Save前钩子函数
func (b *BaseModel[T]) BeforeSave(tx *gorm.DB) (err error) {
	if b.Ctx != nil {
		// 自动维护更新人信息
		userInfo, exise := b.Ctx.Get("identity")
		if exise {
			userInfo := userInfo.(map[string]any)
			if b.Id == 0 { // 新建
				b.CreateBy = userInfo["UserName"].(string)
				b.CreateByName = userInfo["NickName"].(string)
			} else { // 更新
				b.UpdateBy = userInfo["UserName"].(string)
				b.UpdateByName = userInfo["NickName"].(string)
			}
			b.OperatorId = userInfo["UserName"].(string)
			b.OperatorName = userInfo["NickName"].(string)
		}
	}
	return nil
}

// 更新后钩子函数
func (b *BaseModel[T]) AfterUpdate(tx *gorm.DB) (err error) {
	return nil
}

// 删除前钩子函数
func (b *BaseModel[T]) BeforeDelete(tx *gorm.DB) (err error) {
	if b.Ctx != nil {
		// 自动维护更新人信息
		userInfo, exise := b.Ctx.Get("identity")
		if exise {
			userInfo := userInfo.(map[string]any)
			b.UpdateBy = userInfo["UserName"].(string)
			b.UpdateByName = userInfo["NickName"].(string)
		}
	}
	// 处理审计日志信息

	return nil
}

// 删除后钩子函数
func (b *BaseModel[T]) AfterDelete(tx *gorm.DB) (err error) {
	// 异步处理审计日志

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
