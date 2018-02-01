/*
 * 说明：用户数据模型
 * 作者：zhe
 * 时间：2018-01-17 22:55
 * 更新：添加模型
 */

package model

import "gopkg.in/mgo.v2/bson"

type User struct {
	Id       bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"` // omitempty值为空时忽略该字段解析
	Account  string        `json:"account"`                           // 建索引
	Password string        `json:"password"`                          //
	Name     string        `json:"name"`                              //
	Age      int           `json:"age"`                               //
	Email    string        `json:"email"`                             //
	Friends  []string      `json:"friends"`                           // 数组
	Address  Address       `json:"address"`                           // 内嵌文档
	// 数据库私有字段
	CreateAt string `json:"create_at" bson:"create_at"`
	ModifyAt string `json:"modify_at" bson:"modify_at"`
	IsDelete bool   `json:"-" bson:"is_delete"`
	DeleteAt string `json:"-" bson:"delete_at"`
}

type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Remark   string `json:"remark"`
}
