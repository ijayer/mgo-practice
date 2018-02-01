/*
 * 说明：User数据库接口
 * 作者：zhe
 * 时间：2018-01-17 22:56
 * 更新：添加相关函数、操作符的Demo
 */

package dao

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gedex/inflector"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"mongodb.golang.com/src/model"
)

// User数据库访问对象
type UserDao struct {
	dao       *Dao     // 数据库访问对象
	ColName   string   // 集合名称
	IndexKeys []string // 索引字段
}

// 初始化UserDao
func NewUserDao(dao *Dao) *UserDao {
	name := reflect.TypeOf(model.User{}).Name()
	name = inflector.Pluralize(name)
	name = strings.ToLower(name)
	return &UserDao{
		dao:       dao,
		ColName:   name,
		IndexKeys: []string{"account"},
	}
}

// CreateDocDemo
func (d *UserDao) CreateDocDemo() error {
	var err error
	for i := 0; i < 100; i++ {
		user := model.User{
			Id:       bson.NewObjectId(),
			Account:  fmt.Sprintf("mongo_%d", i),
			Password: "123456",
			Name:     "zhe",
			Age:      i,
			Email:    fmt.Sprintf("30%s18@qq.com", RandomMath()),
			Friends:  []string{"KB", "JMS", "YM", "WD", "WS", "KD"},
			Address: model.Address{
				Province: "zj",
				City:     "hz",
				District: "gs",
				Remark:   "Earth",
			},
			CreateAt: Now(),
			ModifyAt: Now(),
			IsDelete: false,
			DeleteAt: "",
		}
		if err = d.dao.CreateDoc(d.ColName, user, d.IndexKeys...); err != nil {
			break
		}
	}
	return err
}

// UpsertDocDemo: 如果文档存在则更新，不存在则创建
// Operators: $set, $setOnInsert
func (d *UserDao) UpsertDocDemo() error {
	user := model.User{
		Account:  fmt.Sprintf("mongo_%s", "a"),
		Password: "123456",
		Name:     "zhe",
		Age:      18,
		Email:    fmt.Sprintf("30%s18@qq.com", RandomMath()),
		Friends:  []string{"KB", "JMS", "YM", "WD", "WS", "KD"},
		Address: model.Address{
			Province: "zj",
			City:     "hz",
			District: "gs",
			Remark:   "Earth",
		},
		ModifyAt: Now(),
		IsDelete: false,
		DeleteAt: "",
	}
	selector := bson.M{"account": user.Account}

	update := bson.M{}
	if err := StructToBsonMap(user, &update); err != nil {
		return err
	}
	delete(update, "create_at")

	// $setOnInsert 设置只在文档创建时需要添加的字段
	change := mgo.Change{
		Update: bson.M{
			"$set":         update,
			"$setOnInsert": bson.M{"create_at": Now(), "is_delete": false, "delete_at": ""},
		},
		Upsert:    true,
		ReturnNew: true,
	}

	changeInfo, err := d.dao.UpsertDoc(d.ColName, selector, change)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", *changeInfo)

	return nil
}

// UpdateDocDemo: 更新文档(部分字段)
// Operators: $set, $unset, $inc, $rename
func (d *UserDao) UpdateDocDemo() error {
	selector := bson.M{"account": "mongo_a"}

	update := bson.M{
		"$set": bson.M{"name": "mongo", "book": "golang", "modify_at": Now()},
		"$inc": bson.M{"age": 6},
	}
	err := d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	// 删除|重命名键
	update = bson.M{
		"$unset":  bson.M{"price": "", "password": ""},
		"$rename": bson.M{"book": "movies"},
	}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateEmbedDocDemo: 更新内嵌文档(整体更新|部分字段更新)
// Operators: $set、.
func (d *UserDao) UpdateEmbedDocDemo() error {
	selector := bson.M{"account": "mongo_a"}

	// 整体更新
	address := model.Address{
		Province: "zhejiang",
		City:     "hangzhou",
		District: "xihu",
		Remark:   "Earth",
	}
	update := bson.M{"$set": bson.M{"address": address, "modify_at": Now()}}
	err := d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	// 部分字段
	update = bson.M{"$set": bson.M{"address.province": "beijing", "modify_at": Now()}}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}
	return nil
}

// UpdateArrDemo: 更新数组(插入|删除元素)
// Operators: $, $pop, $push, $each, $slice, $addToSet
func (d *UserDao) UpdateArrDemo() error {
	selector := bson.M{"account": "mongo_a"}

	// 添加一个元素
	var err error
	var update = make(bson.M)
	update = bson.M{"$push": bson.M{"friends": "You"}}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	// 添加多个元素
	update = bson.M{"$push": bson.M{"friends": bson.M{
		"$each":  []string{"You", "A", "B", "C", "D", "E", "F", "G"}, // 注：这个地方会插入重复数据：You
		"$slice": -5,                                                 // 限定数组长度，且不超过10，超过则保留最后10个
	}}}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	// 查看结果
	results, err := d.dao.FindDoc(d.ColName, selector, Page{})
	if err != nil {
		return err
	}
	BsonMapToJson(results)

	// 删除元素
	update = bson.M{"$pop": bson.M{"friends": -1}} // 从头删除
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	update = bson.M{"$pop": bson.M{"friends": 1}} // 从尾删除
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	// 添加多个不重复元素
	update = bson.M{"$addToSet": bson.M{"friends": bson.M{
		"$each":  []string{"D", "E", "You", "Zhe", "Me"},
		"$slice": 5,
	}}}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}

	return nil
}

// 查询文档
func FindDocDemo() error {
	return nil
}

// 查询数组
func FindArrDemo() error {
	return nil
}

// 模糊查询
func FuzzySearch() error {
	return nil
}

// 聚合查询
func PipeSearchDemo() error {
	return nil
}
