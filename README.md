# :tada: base 业务模型基础业务能力的赋能封装

## :cake: 项目地址
https://github.com/jianyuezhexue/base

## :cake: 项目背景
&nbsp;&nbsp;&nbsp;&nbsp; 日常开发过程中,会发现不同的业务模型,每个都需要重写一部分基础函数,以及在面对复杂需求的场景,面向过程的的开发思路,会让代码随着业务复杂度的增加，变得难以维护。所以，做这个封装有三个基本目的，第一个是通过封装，尝试结构性的矫正面向对象的开发思路；第二个是通过封装和继承，让每个业务模型马上就有基础能力,同时提升开发效率和保证开发质量；第三个是通过这个封装,抽离出一个公共空间,在未来,通过封装实现更多一劳永逸的代码,持续精益开发过程和沉淀高质量代码；

## :cake: 领域和业务建模共识
1. 做好领域划分
1. 每个领域有多个业务模型
1. 每个业务模型有多个能力 | 基础能力 + 业务能力
1. Logic层代码实例化多个业务模型,编排不同的业务能力,从而实现业务逻辑

## :cake: 代码执行标准化流程(SOP)
### input方向
1. 接参&基础校验
1. 实例化实体对象 NewEntity
1. 填充数据 SetData | LoadById(更新场景)
1. 通用逻辑校验 Validate
1. 数据修复 Repair
1. 查询旧数据 GetById
1. 涉及旧数据的校验 | 比如某种状态下不允许修改...
1. 保存或更新数据
1. 记录操作日志

### output方向
1. 接参
1. 实例化实体对象 NewEntity
1. 组合查询条件 MakeCondition
1. 查询数据 LoadData || List
1. 数据完善 Compire
1. 返回数据

### 复杂业务逻辑
1. 基于input方向的节本结构|重点描述逻辑编排部分
1. 实例化实体对象 NewEntity(l.Ctx) | 可以实例化多个模型
1. 设置数据或加载数据 SetData | LoadById
1. 调用模型能力A entity.AblitityA()
1. 开启事务 | entity.Transaction(func(){})
1. 调用模型能力B entity.AblitityB()
1. 事务结束 | 事务失败会自动回滚


## :cake: 基础能力介绍
```go
// 充血模型基础接口
type BaseModelInterface[T any] interface {
	TableName() string                                                                                                               // 表名
	Tx() *gorm.DB                                                                                                                    // 获取事务DB
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error                                                            // 事务处理
	SetData(data any) (*T, error)                                                                                                    // 设置数据
	Validate() error                                                                                                                 // 数据校验
	Create() (*T, error)                                                                                                             // 新增数据
	Update() (*T, error)                                                                                                             // 更新数据
	LoadData(cond SearchCondition, preloads ...PreloadsType) (*T, error)                                                             // 加载数据
	LoadById(id uint64, preloads ...PreloadsType) (*T, error)                                                                        // 根据Id加载数据
	LoadByBusinessCode(filedName, filedValue string, preloads ...PreloadsType) (*T, error)                                           // 根据业务编码查询数据
	GetById(Id uint64, preloads ...PreloadsType) (*T, error)                                                                         // 根据Id查询数据
	GetByIds(Ids []uint64, preloads ...PreloadsType) ([]*T, error)                                                                   // 根据Id查询数据
	Repair() error                                                                                                                   // 修复数据
	Count(conds ...SearchCondition) (int64, error)                                                                                   // 统计数据条数
	List(conds ...SearchCondition) ([]*T, error)                                                                                     // 查询列表数据
	Complete() error                                                                                                                 // 完善数据
	Del(ids ...uint64) error                                                                                                         // 删除数据
	CheckBusinessCodeRepeat(filedName, businessCode string) (bool, error)                                                            // 检查业务编码是否重复
	CheckBusinessCodesExist(filedName string, values []string, more ...SearchCondition) (map[int]bool, error)                        // 批量检查业务编码是否存在
	CheckUniqueKeysRepeat(filedNames []string, values []string, withOutIds ...uint64) (bool, error)                                  // 检查唯一键是否重复
	CheckUniqueKeysRepeatBatch(filedNames []string, values [][]string, withOutIds ...uint64) ([]bool, error)                         // 批量检查唯一键是否重复
	MakeConditon(data any) func(db *gorm.DB) *gorm.DB                                                                                // 构造查询条件
	ReInit(baseModel *BaseModel[T]) error                                                                                            // 重置模型中的Context和Db
	InitStateMachine(initStatus string, events []fsm.EventDesc, afterEvent fsm.Callback, callbacks ...map[string]fsm.Callback) error // 初始化状态机
	EventExecution(initStatus, event, eventZhName string) error                                                                      // 执行事件
}

```

## :cake: 使用示例
```go
// 1. 引入包
import "github.com/jianyuezhexue/base"

// 2.定义业务模型并将baseModel组合进去
type UserEntity struct {
    base.BaseModel[UserEntity]
    Xxx sting `json:"xxx"` // 业务字段
}

// 3.定义模型接口并将baseModel接口组合进去
type UserEntityInterface interface {
    base.BaseModelInterface
    Xxx() error // 更多的模型函数
}

// 4.实现三个基础函数 | Validate() error ，Repair() error，Complete() error

// ValidateFunc 数据校验
func (m *UserEntity) Validate() error {
	// 自定义数据校验逻辑

	// 校验地址是否存在

	// 校验客户Code是否存在

	// more...

	return nil
}

// Repair 数据修复
func (m *UserEntity) Repair() error {
	// 自定义数据修复逻辑

	// 单据来源默认为1

	// EDI对接接口，自动补全用户发票地址

	// more...

	return nil
}

// Complete 数据完善
func (m *UserEntity) Complete() error {
	// 自定义完善数据逻辑

	// 数据字典对应中文名称补全

	return nil
}

// 5. 实现实例化模型函数
// 实例化实体业务模型
func NewUserEntity(ctx *gin.Context, opt ...base.Option[UserEntity]) UserEntityInterface {
	entity := &UserEntity{}
	entity.BaseModel = base.NewBaseModel(ctx, InitDb(), entity.TableName(), entity)

	// 自定义配置选项
	if len(opt) > 0 {
		for _, fc := range opt {
			fc(&entity.BaseModel)
		}
	}
	return entity
}


// 6. 使用示例

// 模拟gin.Context
ctx := &gin.Context{Request: &http.Request{}}
ctx.Set("currUserId", "1")
ctx.Set("currUserName", "张三")

// 实例化业务实体
userEntity := salesOrder.NewUserEntity(ctx)

// 调用基础模型能力
userEntity.SetData(map[string]interface{}{"xxx": "xxx"})   // 设置数据
userEntity.Validate()                                      // 数据校验
userEntity.LoadById(1)                                     // 根据Id加载数据
userEntity.Create()                                        // 新增数据
userEntity.Update()                                        // 更新数据
userEntity.LoadData(base.MakeCondition())                  // 加载数据
userEntity.LoadByBusinessCode("xxx", "xxx")                // 根据业务编码查询数据
userEntity.GetById(1)                                      // 根据Id查询数据
userEntity.GetByIds([]uint64{1, 2})                        // 根据Id查询数据
userEntity.Count()                                         // 统计数据条数
userEntity.List()                                          // 查询列表数据
userEntity.Del(1)                                          // 删除数据
userEntity.Transaction(func(tx *gorm.DB) error {           // 开启事务
    // 事务内操作
    return nil
})
// 更多请参考 exampleLogic/ order_test.go

// 调用自定义能力
userEntity.Xxx() // 自定义业务能力

```

