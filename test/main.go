package main

import (
	"Languege/gorm_wrapper/models"
	"fmt"
	"flag"
	"Languege/redis_wrapper"
	"Languege/gorm_wrapper"
)
var(
	DSN	string
	Debug bool

	RedisIP, RedisPort, RedisPassword	string
)
/**
 *@author LanguageY++2013
 *2019/2/25 3:32 PM
 **/
func init() {
	flag.StringVar(&DSN, "dsn", "", "-dsn=	# for connection")
	flag.BoolVar(&Debug, "debug", false, "-debug=	# print sql for debug")
	flag.StringVar(&RedisIP, "redis_ip", "", "-redis_ip=	 ")
	flag.StringVar(&RedisPort, "redis_port", "", "-redis_port=	")
	flag.StringVar(&RedisPassword, "redis_password", "", "-redis_password=	")

	flag.Parse()
}

func main() {

	//初始化缓存
	redis_wrapper.InitConnect(RedisIP, RedisPort, RedisPassword)
	//初始化db
	gorm_wrapper.Init(DSN, Debug)


	m := &models.UserMail{}
	err := m.FindByPK(1)

	fmt.Println(m, err)



}