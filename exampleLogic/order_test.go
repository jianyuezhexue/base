package examplelogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyuezhexue/base/exampleDomain/salesOrder"
	"github.com/jianyuezhexue/base/exampleDomain/salesOrderDetail"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 新增接口
func TestCreate(t *testing.T) {
	// 0. 模拟数据
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Set("currUserId", "110")
	ctx.Set("currUserName", "张三")

	// 模拟请求数据
	reqData := &salesOrder.CreateSalesOrder{
		OrderId:      fmt.Sprintf("SO%d", time.Now().UnixMicro()),
		CustomerName: "张三",
		Address:      "北京市朝阳区",
		SalesOrderDetails: []*salesOrderDetail.CreateSalesOrderDetail{
			{
				SkuCode:       "SKU001",
				ProductName:   "商品名称",
				BrandName:     "品牌名称",
				ModelType:     "型号",
				OrderQuantity: 1,
			},
		},
	}

	// -------------- 逻辑代码 ---------------

	// 1. 实例化业务实体
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx)

	// 2. 设置数据
	_, err := salesOrderEntity.SetData(reqData)
	assert.Nil(t, err)

	// 3. 校验数据
	err = salesOrderEntity.Validate()
	assert.Nil(t, err)

	// 4. 数据修复
	err = salesOrderEntity.Repair()
	assert.Nil(t, err)

	// 5. 开启事务
	err = salesOrderEntity.Transaction(func(tx *gorm.DB) error {

		// 1. 新增数据
		_, err2 := salesOrderEntity.Create()
		if err2 != nil {
			return err2
		}

		// 2. more...

		return nil
	})
	assert.Nil(t, err)

}

// 列表接口
func TestList(t *testing.T) {

	// 0. 模拟数据
	ctx := &gin.Context{}
	ctx.Set("currUserId", "110")
	ctx.Set("currUserName", "张三")

	// 模拟请求数据
	reqData := salesOrder.SearchSalesOrder{
		Id:       1,
		Page:     1,
		PageSize: 10,
		CreatedAt: []string{
			time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			time.Now().Format("2006-01-02 15:04:05"),
		},
		CustomerName:     "张三",
		CustomerNameLike: "张",
	}

	// -------------- 逻辑代码 ---------------

	// 1. 实例化业务实体
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx)

	// 2. 组合搜索条件 | 注意这里传入的不能是指针
	condtion := salesOrderEntity.MakeConditon(reqData)

	// 3. 查询分页
	total, err := salesOrderEntity.Count(condtion)
	assert.Nil(t, err)

	// 4. 查询分页数据
	list, err := salesOrderEntity.List(condtion)
	assert.Nil(t, err)

	// 5. 对数据进行完善
	for _, item := range list {
		err = item.Complete()
		assert.Nil(t, err)
	}

	// 6. 组合返回数据
	resp := &salesOrder.ListReap{
		Page:     reqData.Page,
		PageSize: reqData.PageSize,
		Total:    total,
		List:     list,
	}

	// 7. 返回数据
	bytes, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Printf(string(bytes))
}
