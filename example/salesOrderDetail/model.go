package salesOrderDetail

import (
	"github.com/jianyuezhexue/base"
)

// SalesOrderDetailEntity 实体业务模型
type SalesOrderDetail struct {
	base.BaseModel[SalesOrderDetailEntity]
	OrderId                   string  `json:"orderId" type:"db" search:"type:eq;column:order_id;table:sales_order_detail" comment:"SO号"`                                                         // SO号
	SkuCode                   string  `json:"skuCode" type:"db" search:"type:eq;column:sku_code;table:sales_order_detail" comment:"SKU编码"`                                                       // SKU编码
	ProductName               string  `json:"productName" type:"db" search:"type:eq;column:product_name;table:sales_order_detail" comment:"SKU名称"`                                               // SKU名称
	BrandName                 string  `json:"brandName" type:"db" search:"type:eq;column:brand_name;table:sales_order_detail" comment:"品牌"`                                                      // 品牌
	ModelType                 string  `json:"modelType" type:"db" search:"type:eq;column:model_type;table:sales_order_detail" comment:"型号"`                                                      // 型号
	OrderQuantity             float64 `json:"orderQuantity" type:"db" search:"type:eq;column:order_quantity;table:sales_order_detail" comment:"订单数量"`                                            // 订单数量
	CancelQuantity            float64 `json:"cancelQuantity" type:"db" search:"type:eq;column:cancel_quantity;table:sales_order_detail" comment:"取消数量"`                                          // 取消数量
	Unit                      string  `json:"unit" type:"db" search:"type:eq;column:unit;table:sales_order_detail" comment:"单位"`                                                                 // 单位
	UnitPrice                 float64 `json:"unitPrice" type:"db" search:"type:eq;column:unit_price;table:sales_order_detail" comment:"单价"`                                                      // 单价
	TotalPrice                float64 `json:"totalPrice" type:"db" search:"type:eq;column:total_price;table:sales_order_detail" comment:"总价"`                                                    // 总价
	CustomerMaterialCode      string  `json:"customerMaterialCode" type:"db" search:"type:eq;column:customer_material_code;table:sales_order_detail" comment:"客户物料编码"`                           // 客户物料编码
	CustomerMaterialName      string  `json:"customerMaterialName" type:"db" search:"type:eq;column:customer_material_name;table:sales_order_detail" comment:"客户物料名称"`                           // 客户物料名称
	GoodsRowRemark            string  `json:"goodsRowRemark" type:"db" search:"type:eq;column:goods_row_remark;table:sales_order_detail" comment:"商品行备注"`                                        // 商品行备注
	SupplyChainRemark         string  `json:"supplyChainRemark" type:"db" search:"type:eq;column:supply_chain_remark;table:sales_order_detail" comment:"供应商备注"`                                  // 供应商备注
	IsDirectDeliveryToStation int     `json:"isDirectDeliveryToStation" type:"db" search:"type:eq;column:is_direct_delivery_to_station;table:sales_order_detail" comment:"直发是否允许进配送站[0-否 ,1-是]"` // 直发是否允许进配送站[0-否 ,1-是]
}

type SalesOrderDetailEntity struct {
	SalesOrderDetail
	ActualQuantity float64 `json:"actualQuantity" gorm:"-" comment:"实际数量"` // 实际数量
}

// 数据表名
func (m *SalesOrderDetailEntity) TableName() string {
	return "sales_order_detail"
}
