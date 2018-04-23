/*
 * 说明：Tutorial for Mongodb based on Golang and MongoDB
 * 作者：zhe
 * 时间：2018-01-17 22:55
 * 更新：
 */

package main

import (
	"fmt"

	"mongodb.golang.com/src/dao"
)

func main() {
	session := dao.InitMongo()
	defer session.Close()

	d := dao.NewDao(session)
	userDao := dao.NewUserDao(d)

	var err error

	err = userDao.TestFindOneResultJsonMarshal()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}

	err = userDao.TestFindAllResultJsonMarshal()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}

	err = userDao.FindDocDemo()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	}
}
