package salesOrderDetail

// 新增销售订单
type CreateSalesOrderDetail struct {
	OrderId                   string  `json:"-"`
	SkuCode                   string  `json:"skuCode" uri:"skuCode" form:"skuCode" vd:"len($)>0 && mblen($)<=10;msg:'SKU编码[必填且字符长度不能超过10]'"`
	ProductName               string  `json:"productName" uri:"productName" form:"productName" vd:"len($)>0 && mblen($)<=100;msg:'SKU名称[必填且字符长度不能超过100]'"`
	BrandName                 string  `json:"brandName" uri:"brandName" form:"brandName" vd:"mblen($)<=100;msg:'品牌[字符长度不能超过20]'"`
	ModelType                 string  `json:"modelType" uri:"modelType" form:"modelType" vd:"mblen($)<=100;msg:'型号[字符长度不能超过20]'"`
	OrderQuantity             float64 `json:"orderQuantity" uri:"orderQuantity" form:"orderQuantity" `
	Unit                      string  `json:"unit" uri:"unit" form:"unit" vd:"mblen($)<=100;msg:'单位[字符长度不能超过10]'"`
	UnitPrice                 float64 `json:"unitPrice" uri:"unitPrice" form:"unitPrice"`
	TotalPrice                float64 `json:"totalPrice" uri:"totalPrice" form:"totalPrice"`
	CustomerMaterialCode      string  `json:"customerMaterialCode" uri:"customerMaterialCode" form:"customerMaterialCode" vd:"mblen($)<=100;msg:'客户物料编码[字符长度不能超过100]'"`
	CustomerMaterialName      string  `json:"customerMaterialName" uri:"customerMaterialName" form:"customerMaterialName" vd:"mblen($)<=100;msg:'客户物料名称[字符长度不能超过100]'"`
	GoodsRowRemark            string  `json:"goodsRowRemark" uri:"goodsRowRemark" form:"goodsRowRemark" vd:"mblen($)<=200;msg:'商品行备注[字符长度不能超过200]'"`
	SupplyChainRemark         string  `json:"supplyChainRemark" uri:"supplyChainRemark" form:"supplyChainRemark" vd:"mblen($)<=200;msg:'供应商备注[字符长度不能超过200]'"` // 供应链备注
	IsDirectDeliveryToStation int     `json:"isDirectDeliveryToStation" uri:"isDirectDeliveryToStation" form:"isDirectDeliveryToStation" `                    // 直发是否允许进配送站 0 -否  ,1-是
}

// 更新销售订单
type UpdateSalesOrderDetail struct {
	Id                   int     `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderId              string  `json:"orderId" gorm:"column:order_id;type:varchar(100);not null;default:''"`                            // SO号
	SkuCode              string  `json:"skuCode" gorm:"column:sku_code;type:varchar(100);not null;default:''"`                            // sku_code
	ProductName          string  `json:"productName" gorm:"column:product_name;type:varchar(100);not null;default:''"`                    // sku名称
	BrandName            string  `json:"brandName" gorm:"column:brand_name;type:varchar(100);not null;default:''"`                        // 品牌
	ModelType            string  `json:"modelType" gorm:"column:model_type;type:varchar(100);not null;default:''"`                        // 型号
	OrderQuantity        string  `json:"orderQuantity" gorm:"column:order_quantity;type:varchar(100);not null;default:''"`                // 订单数量
	Unit                 string  `json:"unit" gorm:"column:unit;type:varchar(100);not null;default:''"`                                   // 单位
	UnitPrice            float64 `json:"unitPrice" gorm:"column:unit_price;type:decimal(10,2);not null;default:0.00"`                     // 单价
	TotalPrice           float64 `json:"totalPrice" gorm:"column:total_price;type:decimal(10,2);not null;default:0.00"`                   // 总价
	CustomerMaterialCode string  `json:"customerMaterialCode" gorm:"column:customer_material_code;type:varchar(100);not null;default:''"` // 客户物料编码
	CustomerMaterialName string  `json:"customerMaterialName" gorm:"column:customer_material_name;type:varchar(100);not null;default:''"` // 客户物料名称
	GoodsRowRemark       string  `json:"goodsRowRemark" gorm:"column:goods_row_remark;type:varchar(250);not null;default:''"`             // 商品行备注
	UpdateBy             uint    `json:"updateBy" gorm:"column:update_by;type:int unsigned;not null;default:0"`                           // 修改人id
	UpdateByName         string  `json:"updateByName" gorm:"column:update_by_name;type:varchar(20);not null;default:'更新人姓名'"`             // 更新人姓名
}

// 搜索销售订单
type SearchSalesOrderDetail struct {
	Id                   int      `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderId              string   `json:"orderId" gorm:"column:order_id;type:varchar(100);not null;default:''" search:"type:eq;column:order_id;table:sales_order_detail"`                                          // SO号
	SkuCode              string   `json:"skuCode" gorm:"column:sku_code;type:varchar(100);not null;default:''" search:"type:eq;column:sku_code;table:sales_order_detail"`                                          // sku_code
	ProductName          string   `json:"productName" gorm:"column:product_name;type:varchar(100);not null;default:''" search:"type:eq;column:product_name;table:sales_order_detail"`                              // sku名称
	BrandName            string   `json:"brandName" gorm:"column:brand_name;type:varchar(100);not null;default:''" search:"type:eq;column:brand_name;table:sales_order_detail"`                                    // 品牌
	ModelType            string   `json:"modelType" gorm:"column:model_type;type:varchar(100);not null;default:''" search:"type:eq;column:model_type;table:sales_order_detail"`                                    // 型号
	OrderQuantity        string   `json:"orderQuantity" gorm:"column:order_quantity;type:varchar(100);not null;default:''" search:"type:eq;column:order_quantity;table:sales_order_detail"`                        // 订单数量
	Unit                 string   `json:"unit" gorm:"column:unit;type:varchar(100);not null;default:''" search:"type:eq;column:unit;table:sales_order_detail"`                                                     // 单位
	UnitPrice            float64  `json:"unitPrice" gorm:"column:unit_price;type:decimal(10,2);not null;default:0.00" search:"type:eq;column:unit_price;table:sales_order_detail"`                                 // 单价
	TotalPrice           float64  `json:"totalPrice" gorm:"column:total_price;type:decimal(10,2);not null;default:0.00" search:"type:eq;column:total_price;table:sales_order_detail"`                              // 总价
	CustomerMaterialCode string   `json:"customerMaterialCode" gorm:"column:customer_material_code;type:varchar(100);not null;default:''" search:"type:eq;column:customer_material_code;table:sales_order_detail"` // 客户物料编码
	CustomerMaterialName string   `json:"customerMaterialName" gorm:"column:customer_material_name;type:varchar(100);not null;default:''" search:"type:eq;column:customer_material_name;table:sales_order_detail"` // 客户物料名称
	GoodsRowRemark       string   `json:"goodsRowRemark" gorm:"column:goods_row_remark;type:varchar(250);not null;default:''" search:"type:eq;column:goods_row_remark;table:sales_order_detail"`                   // 商品行备注
	CreatedAt            []string `json:"createdAt" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP" search:"type:between;column:created_at;table:sales_order_detail"`                    // 创建时间
	UpdatedAt            []string `json:"updatedAt" gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP" search:"type:between;column:updated_at;table:sales_order_detail"`                             // 更新时间
	Page                 int64    `json:"page" gorm:"-" search:"page"`                                                                                                                                             // 分页
	PageSize             int64    `json:"pageSize" gorm:"-" search:"pageSize"`                                                                                                                                     // 分页大小
}

// 删除销售订单
type DelSalesOrderDetail struct {
	Ids []uint64 `json:"ids"`
}
