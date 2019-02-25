package gorm_wrapper

import (
	"github.com/jinzhu/gorm"
	"flag"
	"Languege/redis_wrapper"
	_ "github.com/go-sql-driver/mysql"
)

/**
 *@author LanguageY++2013
 *2019/2/25 2:14 PM
 **/
var(
	DB *gorm.DB
	DSN	string
	Debug bool

	RedisIP, RedisPort, RedisPassword	string
)

func init() {
	flag.StringVar(&DSN, "dsn", "", "-dsn=	# for connection")
	flag.BoolVar(&Debug, "debug", false, "-debug=	# print sql for debug")
	flag.StringVar(&RedisIP, "redis_ip", "", "-redis_ip=	 ")
	flag.StringVar(&RedisPort, "redis_port", "", "-redis_port=	")
	flag.StringVar(&RedisPassword, "redis_password", "", "-redis_password=	")

	flag.Parse()

	var err error
	DB, err = gorm.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}

	if Debug {
		DB.LogMode(true)
	}

	//初始化缓存
	redis_wrapper.InitConnect(RedisIP, RedisPort, RedisPassword)
}

