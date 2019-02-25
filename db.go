package gorm_wrapper

import (
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
)

/**
 *@author LanguageY++2013
 *2019/2/25 2:14 PM
 **/
var(
	DB *gorm.DB
)


func Init(dsn string, debug bool) {
	var err error
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if debug {
		DB.LogMode(true)
	}
}

