package mysql

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB

func Conn(dns string) error {
	var err error
	Db, err = gorm.Open(mysql.Open(dns), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}

	sqlDB, dbErr := Db.DB()
	if dbErr != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(100)

	return nil
}

func CreateTable(tableName string, fields []string) error {
	var err error
	delSql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	err = Db.Exec(delSql).Error
	if err != nil {
		return err
	}

	s := "id bigint(20) NOT NULL PRIMARY KEY"
	for _, field := range fields {
		s += fmt.Sprintf(",%s varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL", field)
	}
	sql := fmt.Sprintf("CREATE TABLE `%s` (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", tableName, s)

	err = Db.Exec(sql).Error
	if err != nil {
		return err
	}

	return nil
}

func InsertData(tableName string, fields []string, data []bson.M) error {
	var err error
	var maps []map[string]interface{}
	for _, d := range data {
		row := make(map[string]interface{})
		for _, field := range fields {
			row[field] = d[field]
		}
		if row != nil {
			row["id"] = d["id"].(string)
			maps = append(maps, row)
		}
	}

	if len(maps) > 0 {
		err = Db.Table(tableName).CreateInBatches(maps, 100).Error
		if err != nil {
			return err
		}
	}

	return err
}
