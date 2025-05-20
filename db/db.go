package db

import (
	"database/sql"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDb() *gorm.DB {
	dataSourceName := "root:root@tcp(localhost:3306)/admin?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
	sqlDB, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic("数据库连接失败:" + err.Error())
	}
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Second * 28800) // SHOW VARIABLES LIKE '%timeout%';

	// 生成gorm连接
	Db, err := gorm.Open(
		mysql.New(mysql.Config{Conn: sqlDB}),
		&gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}},
	)
	if err != nil {
		panic(err)
	}
	return Db
}
