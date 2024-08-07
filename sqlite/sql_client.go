package sqlite

import (
	"database/sql"
	"fmt"
	"go_code/myselfgo/inits"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDB() (*sql.DB, error) {
	Db := inits.Config.MysqlClient
	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", Db.User, Db.Password, Db.Host, Db.Port, Db.DbName)
	// 打开数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, err
	}
	// 设置数据库连接池的最大连接数
	db.SetMaxOpenConns(10)
	return db, nil
}

func InitOrm() (*gorm.DB, error) {
	Db := inits.Config.MysqlClient
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", Db.User, Db.Password, Db.Host, Db.Port, Db.DbName)
	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名单数形式
		},
		// 可以在这里添加更多GORM配置选项
	})
	if err != nil {
		return nil, err
	}
	// 自动迁移schema（可选）
	// if err := orm.AutoMigrate(&YourModel1{}, &YourModel2{}); err != nil {
	// 	return nil, err
	// }
	return orm, nil
}
