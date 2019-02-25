package main

import (
	"Languege/gorm_wrapper/models"
	"fmt"
)

/**
 *@author LanguageY++2013
 *2019/2/25 3:32 PM
 **/
func main() {

	m := &models.UserMail{}
	err := m.FindByPK(1)

	fmt.Println(m, err)



}