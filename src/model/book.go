/*
 * 说明：
 * 作者：zhe
 * 时间：2018-02-01 10:50
 * 更新：
 */

package model

import "gopkg.in/mgo.v2"

type Book struct {
	Name    string    `json:"name"`
	Price   float64   `json:"price"`
	Author  string    `json:"author"`
	UserRef mgo.DBRef `json:"user_ref"`
}
