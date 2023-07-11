package src

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var controller = NewController(&dbMySQL)

func TestInitTable(t *testing.T) {
	initMySQLDB()
	controller.Init()
}

func TestFileNameCheck(t *testing.T) {
	controller.QuerySqlFiles("../sql/")
}

func TestHashCheck(t *testing.T) {
	fmt.Println(Hash("Hello world..."))
}

func TestFilesCheck(t *testing.T) {
	controller.QueryAllVersion()
	controller.ReadFilesInDisk("../sql/")
	controller.CheckSqlFiles()
	controller.ExecuteSqlFiles()
}

var dbMySQL = gorm.DB{}

func initMySQLDB() {
	var datetimePrecision = 2
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root:123@qq.com@tcp(127.0.0.1:3306)/go-mysql-version-control?charset=utf8&parseTime=True&loc=Local", // data source name, refer https://github.com/go-sql-driver/mysql#dsn-data-source-name
		DefaultStringSize:         256,                                                                                                  // add default size for string fields, by default, will use db type `longtext` for fields without size, not a primary key, no index defined and don't have default values
		DisableDatetimePrecision:  true,                                                                                                 // disable datetime precision support, which not supported before MySQL 5.6
		DefaultDatetimePrecision:  &datetimePrecision,                                                                                   // default datetime precision
		DontSupportRenameIndex:    true,                                                                                                 // drop & create index when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                                                                                                 // use change when rename column, rename rename not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                                                                                                // smart configure based on used version
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logrus.Error("connect db fail:", err.Error())
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	dbMySQL = *db
}
