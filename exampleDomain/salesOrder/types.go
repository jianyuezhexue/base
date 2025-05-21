package salesOrder

import "github.com/jianyuezhexue/base/exampleDomain/salesOrderDetail"

// 新增销售订单
type CreateSalesOrder struct {
	OrderId           string                                     `json:"orderId" uri:"orderId" form:"orderId" vd:"len($)>0&&mblen($)<=40;msg:'SO号[必填,字符长度不超过40]'"`                                       // SO号
	CustomerName      string                                     `json:"customerName" uri:"customerName" form:"customerName" vd:"len($)>0&&mblen($)<=100;msg:'客户名称[必填,字符长度不超过100]'"`                     // 客户名称
	Address           string                                     `json:"address" uri:"address" form:"address" vd:"len($)>0&&mblen($)<=200;msg:'地址[必填,字符长度不超过200]'"`                                      // 地址
	SalesOrderDetails []*salesOrderDetail.CreateSalesOrderDetail `json:"salesOrderDetails" type:"realtion" gorm:"foreignKey:ShippingId;references:Id;" comment:"销售单明细" vd:"len($)>0;msg:'销售单明细[必须有一条]'"` // 销售单明细
}

// 更新销售订单
type UpdateSalesOrder struct {
	Id                int64                                      `json:"id" uri:"id" form:"id"`
	OrderId           string                                     `json:"orderId" uri:"orderId" form:"orderId" vd:"mblen($)<=100;msg:'SO号[字符长度不超过100]'"`                                                  // SO号
	CustomerName      string                                     `json:"customerName" uri:"customerName" form:"customerName" vd:"len($)>0&&mblen($)<=100;msg:'客户名称[必填,字符长度不超过100]'"`                     // 客户名称
	Address           string                                     `json:"address" uri:"address" form:"address" vd:"len($)>0&&mblen($)<=200;msg:'地址[必填,字符长度不超过200]'"`                                      // 地址
	SalesOrderDetails []*salesOrderDetail.CreateSalesOrderDetail `json:"salesOrderDetails" type:"realtion" gorm:"foreignKey:ShippingId;references:Id;" comment:"销售单明细" vd:"len($)>0;msg:'销售单明细[必须有一条]'"` // 销售单明细
}

// 搜索销售订单
type SearchSalesOrder struct {
	Id               uint64   `json:"id" search:"type:eq;column:id;table:sales_order"`                            // 主键ID
	Page             int64    `json:"page" search:"page"`                                                         // 分页
	PageSize         int64    `json:"pageSize"  search:"pageSize"`                                                // 分页大小
	CreatedAt        []string `json:"createdAt" search:"type:between;column:created_at;table:sales_order"`        // 创建时间
	UpdatedAt        []string `json:"updatedAt" search:"type:between;column:updated_at;table:sales_order"`        // 更新时间
	OrderId          string   `json:"orderId"  search:"type:eq;column:order_id;table:sales_order" `               // SO号
	CustomerName     string   `json:"customerName" search:"type:eq;column:customer_name;table:sales_order"`       // 客户名称
	CustomerNameLike string   `json:"customerNameLike" search:"type:like;column:customer_name;table:sales_order"` // 客户名称-like
	Address          string   `json:"address" search:"type:eq;column:address;table:sales_order"`                  // 地址
}

type ListReap struct {
	Page     int64               `json:"page" comment:"页数"`
	PageSize int64               `json:"pageSize" comment:"每页数量"`
	Total    int64               `json:"total" comment:"总条数"`
	List     []*SalesOrderEntity `json:"list" comment:"数据"`
}
