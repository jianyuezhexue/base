package salesOrderDetail

import (
	"github.com/jianyuezhexue/base"
)

type SalesOrderDetailInterface interface {
	base.BaseModelInterface[SalesOrderDetailEntity]
}

type SalesOrderDetailEntity struct {
	base.BaseModel[SalesOrderDetailEntity]
	OrderId       string  `json:"orderId"  comment:"SO号"`       // SO号
	SkuCode       string  `json:"skuCode" comment:"SKU编码"`      // SKU编码
	ProductName   string  `json:"productName" comment:"SKU名称"`  // SKU名称
	BrandName     string  `json:"brandName" comment:"品牌"`       // 品牌
	ModelType     string  `json:"modelType" comment:"型号"`       // 型号
	OrderQuantity float64 `json:"orderQuantity" comment:"订单数量"` // 订单数量
}

// 数据表名
func (m *SalesOrderDetailEntity) TableName() string {
	return "sales_order_detail"
}
