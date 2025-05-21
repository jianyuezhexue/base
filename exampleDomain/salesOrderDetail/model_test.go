package salesOrderDetail

import (
	"testing"

	"github.com/jianyuezhexue/base"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 链接DB
func dbConnect() *gorm.DB {
	// 初始化数据库连接
	dsn := "root:root@tcp(localhost:3306)/admin?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// Demo
func TestSalesOrderDetailGetById(t *testing.T) {
	salesOrderDetail := &SalesOrderDetailEntity{}
	salesOrderDetail.BaseModel = base.BaseModel[SalesOrderDetailEntity]{Db: dbConnect()}
	res, err := salesOrderDetail.GetById(1)
	assert.Nil(t, err)
	t.Log(res)
}
