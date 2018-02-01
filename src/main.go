/*
 * 说明：Tutorial for Mongodb based on Golang and Mgo
 * 作者：zhe
 * 时间：2018-01-17 22:55
 * 更新：
 */

package main

import (
	"log"

	"mongodb.golang.com/src/dao"
)

func main() {
	session := dao.InitMongo()
	defer session.Close()

	d := dao.NewDao(session)
	userDao := dao.NewUserDao(d)

	if err := userDao.UpdateArrDemo(); err != nil {
		log.Println(err)
	}
}
