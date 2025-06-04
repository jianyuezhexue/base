package examplelogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyuezhexue/base"
	"github.com/jianyuezhexue/base/exampleDomain/salesOrder"
	"github.com/jianyuezhexue/base/exampleDomain/salesOrderDetail"
	"github.com/looplab/fsm"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 新增接口
func TestCreate(t *testing.T) {
	// 0. 模拟数据
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Set("currUserId", "1")
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

// 更新接口
func TestUpdate(t *testing.T) {
	// 0. 模拟数据
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Set("currUserId", "2")
	ctx.Set("currUserName", "李四")

	// 模拟请求数据
	reqData := &salesOrder.UpdateSalesOrder{
		Id:           1,
		OrderId:      fmt.Sprintf("SO%d", time.Now().UnixMicro()),
		CustomerName: "张三2",
		Address:      "北京市朝阳区2",
		SalesOrderDetails: []*salesOrderDetail.UpdateSalesOrderDetail{
			{
				Id:            1,
				SkuCode:       "SKU002",
				ProductName:   "商品名称2",
				BrandName:     "品牌名称2",
				ModelType:     "型号2--**--",
				OrderQuantity: 2 * 2,
			},
		},
	}

	// -------------- 逻辑代码 ---------------

	// 1. 实例化业务实体
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx)

	// 2. 查询数据
	preloads := map[string][]any{"SalesOrderDetails": {}}
	_, err := salesOrderEntity.LoadById(reqData.Id, preloads)
	assert.Nil(t, err)

	// 3. 设置数据
	_, err = salesOrderEntity.SetData(reqData)
	assert.Nil(t, err)

	// 4. 校验数据
	err = salesOrderEntity.Validate()
	assert.Nil(t, err)

	// 5. 数据修复
	err = salesOrderEntity.Repair()
	assert.Nil(t, err)

	// 6. 开启事务
	err = salesOrderEntity.Transaction(func(tx *gorm.DB) error {

		// 1. 新增数据
		_, err2 := salesOrderEntity.Update()
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
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Set("currUserId", "110")
	ctx.Set("currUserName", "张三")

	// 模拟请求数据
	reqData := salesOrder.SearchSalesOrder{
		Page:             1,
		PageSize:         10,
		CustomerNameLike: "张",
	}

	// -------------- 逻辑代码 ---------------
	// 1. 实例化业务实体
	preloads := map[string][]any{"SalesOrderDetails": {}}
	withPreloads := base.WithPreloads[salesOrder.SalesOrderEntity](preloads)
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx, withPreloads)

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
	fmt.Printf("%s", string(bytes))
}

// TestCheckUniqueKeysExistBatch
func TestCheckUniqueKeysExistBatch(t *testing.T) {
	ctx := &gin.Context{Request: &http.Request{}}
	// 1. 实例化业务实体
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx)

	// 2. 组合数据
	uniqueKeys := []string{"order_id", "customer_name"}
	uniqueValues := [][]string{
		{"SO1111111111111111111", "张三"},
		{"SO1748435979939020", "张三2"},
	}

	// 3. 校验
	repeats, err := salesOrderEntity.CheckUniqueKeysExistBatch(uniqueKeys, uniqueValues)
	assert.Nil(t, err)
	assert.False(t, repeats[0])
	assert.True(t, repeats[1])
}

// TestEventExecution
// 定义事件
// 假设订单状态枚举 0-制单 1-确认订单 2-部分发货 3-全部发货 4-签收 5-回单 6-退货
var Events = []fsm.EventDesc{
	// 制单 ---确认--> 确认订单
	{Src: []string{"0"}, Name: "confirm", Dst: "1"},
	// 确认订单,部分发货 ---部分发货--> 部分发货
	{Src: []string{"1", "2"}, Name: "partDelivery", Dst: "2"},
	// 确认订单,部分发货 ---全部发货--> 全部发货
	{Src: []string{"1", "2"}, Name: "allDelivery", Dst: "3"},
	// 全部发货 ---签收--> 签收
	{Src: []string{"3"}, Name: "signFor", Dst: "4"},
	// 签收 ---回单--> 回单
	{Src: []string{"4"}, Name: "back", Dst: "5"},
	// 签收,回单 ---退货--> 退货
	{Src: []string{"4", "5"}, Name: "returnGoods", Dst: "6"},
}

func TestEventExecution(t *testing.T) {
	// 0. 模拟数据
	ctx := &gin.Context{Request: &http.Request{}}
	ctx.Set("currUserId", "110")
	ctx.Set("currUserName", "张三")

	// 1. 实例化业务实体
	salesOrderEntity := salesOrder.NewSalesOrderEntity(ctx)

	// 2. 根据ID查询数据
	saleOrderData, err := salesOrderEntity.LoadById(1)
	assert.Nil(t, err)

	// 3. 初始化状态机
	err = salesOrderEntity.InitStateMachine(strconv.Itoa(saleOrderData.Status), Events, saleOrderData.EventCallBack)
	assert.Nil(t, err)

	// 4. 开启事务
	err = salesOrderEntity.Transaction(func(tx *gorm.DB) error {
		// 1. 更新状态为确认订单
		saleOrderData.Status = 2
		_, err2 := salesOrderEntity.Update()
		if err2 != nil {
			return err2
		}

		// 2. 执行全部发货事件
		err2 = salesOrderEntity.EventExecution(strconv.Itoa(saleOrderData.Status), "allDelivery", "全部发货")
		if err2 != nil {
			return err2
		}

		return nil
	})
	assert.Nil(t, err)

	// 5. 重新查库校验状态
	saleOrderData, err = salesOrderEntity.LoadById(1)
	assert.Nil(t, err)
	assert.Equal(t, 3, saleOrderData.Status)
}
