package salesOrder

import (
	"github.com/gin-gonic/gin"
	"github.com/jianyuezhexue/base"
	"github.com/jianyuezhexue/base/db"
	"github.com/jianyuezhexue/base/example/salesOrderDetail"
)

// 业务模型接口定义
type SalesOrderInterface interface {
	base.BaseModelInterface[SalesOrderEntity]
}

// 业务模型实体
type SalesOrderEntity struct {
	base.BaseModel[SalesOrderEntity]
	OrderId           string                                     `json:"orderId" comment:"订单号"`                                                                           // SO号
	Status            int                                        `json:"status"  comment:"订单状态"`                                                                          // 订单状态
	SalesOrderDetails []*salesOrderDetail.SalesOrderDetailEntity `json:"salesOrderDetails" type:"realtion" gorm:"foreignKey:OrderId;references:OrderId;" comment:"销售单明细"` // 发货单详情
}

// 数据表名
func (m *SalesOrderEntity) TableName() string {
	return "sales_order"
}

// 实例化实体业务模型
func NewSalesOrderEntity(ctx *gin.Context, opt ...base.Option[SalesOrderEntity]) SalesOrderInterface {
	entity := &SalesOrderEntity{}
	entity.BaseModel = base.NewBaseModel(ctx, db.InitDb(), entity.TableName(), entity)

	// 自定义配置选项
	if len(opt) > 0 {
		for _, fc := range opt {
			fc(&entity.BaseModel)
		}
	}

	return entity
}

// Repair 数据修复
func (m *SalesOrderEntity) Repair() error {
	// 自定义数据修复逻辑

	return nil
}

// Complete 数据完善
func (m *SalesOrderEntity) Complete() error {
	// 自定义完善数据逻辑

	return nil
}

// more abilits...
