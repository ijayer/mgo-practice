/*
 * 说明：User数据库接口
 * 作者：zhe
 * 时间：2018-01-17 22:56
 * 更新：添加相关函数、操作符的Demo
 */

package dao

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	for i := 0; i < 9; i++ {
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
		Comments: []model.Comment{},
		ModifyAt: Now(),
		IsDelete: false,
		DeleteAt: "",
	}
	selector := bson.M{"account": user.Account}

	update := bson.M{}
	if err := StructToMap(user, &update); err != nil {
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
		"$slice": -5,                                                 // 限定数组长度;且不超过10;超过则保留最后10个
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

// UpdateEmbedArrDocDemo: 更新内嵌数组文档
// Operators:
func (d *UserDao) UpdateEmbedArrDocDemo() error {
	selector := bson.M{"account": "mongo_a"}
	user, err := d.dao.FindOneDoc(d.ColName, selector)
	if err != nil {
		return err
	}
	result := user.(bson.M)

	userRef := mgo.DBRef{
		Collection: d.ColName,
		Id:         result["_id"].(bson.ObjectId),
		Database:   d.dao.Name,
	}

	comments := model.Comment{
		Id:       bson.NewObjectId(),
		Content:  "Code compile",
		UserRef:  userRef,
		CreateAt: Now(),
		ModifyAt: Now(),
		IsDelete: false,
		DeleteAt: "",
	}
	err = d.dao.UpdateDoc(d.ColName, selector, bson.M{"$push": bson.M{"comments": comments}})
	if err != nil {
		return err
	}
	return nil
}

// 查询文档
func (d *UserDao) FindDocDemo() error {
	page := Page{}
	page.checkValid("0", "5")
	sortKeys := []string{"-age"}

	var err error
	var result interface{}

	// 按嵌入文档字段查询
	condition := bson.M{"account": "mongo_0"}
	result, err = d.dao.FindDoc(d.ColName, condition, page, sortKeys...)
	if err != nil {
		return err
	}
	fmt.Printf("type: %T, data:%+v\n", result, result)

	results := result.([]bson.M)
	fmt.Println("total:", len(results), "data:", results)

	return nil
}

// 查询文档：指定需要的字段
func (d *UserDao) FindWithSelectDemo() error {
	session := d.dao.SessionCopy()
	defer session.Close()
	co := d.dao.GetCollection(d.ColName, session)

	var q *mgo.Query
	query := bson.M{"age": 2}
	q = co.Find(query)

	var results []interface{}

	// 返回所有字段
	selector := bson.M{}
	err := q.Select(selector).All(&results)
	if err != nil {
		return err
	}
	BsonMapToJson(results...)

	// 只返回指定值为1的字段
	selector = bson.M{"name": 1, "age": 1}
	err = q.Select(selector).All(&results)
	if err != nil {
		return err
	}
	BsonMapToJson(results...)

	// 指定为 0 的字段都不返回;其余都返回
	selector = bson.M{"name": 0, "age": 0}
	err = q.Select(selector).All(&results)
	if err != nil {
		return err
	}
	BsonMapToJson(results...)

	// {"name": 1, "age": 0} 是互斥操作
	// output err: Projection cannot have a mix of inclusion and exclusion.
	selector = bson.M{"name": 1, "age": 0}
	return q.Select(selector).All(&results)
}

// 查询&修改数组、内嵌数组文档
// Operators: $all
func (d *UserDao) FindEmbedArrDemo() error {
	page := Page{}
	page.checkValid("0", "2")

	// 查询数组中包含所有指定元素的文档
	query := bson.M{"friends": bson.M{"$all": []string{"KD", "YM"}}}
	results, err := d.dao.FindDoc(d.ColName, query, page)
	if err != nil {
		return err
	}
	BsonMapToJson(results)

	// 更新内嵌数组文档
	selector := bson.M{
		"account":  "mongo_a",
		"comments": bson.M{"$elemMatch": bson.M{"content": "This is a comment"}},
	}
	update := bson.M{"$set": bson.M{ // $ 操作符最后会取代满足 selector 条件的第一个数据元素对应的 index
		"comments.$.email":     "303xx680@qq.com",
		"comments.$.stars":     6,
		"comments.$.modify_at": Now(),
		"modify_at":            Now(),
	}}
	err = d.dao.UpdateDoc(d.ColName, selector, update)
	if err != nil {
		return err
	}
	return nil
}

// 模糊查询: 关键字查询
// 只匹配可能满足正则规则的某(几)个字段;并不是进行全文匹配
// 函数调用实例：
/*
	if err := userDao.FuzzySearch("zhe1"); err != nil {
		fmt.Println(err)
	}
*/
func (d *UserDao) FuzzySearch(keys ...string) error {
	condition := bson.M{}
	ms := MatchKeys(keys...)
	if len(ms) == 0 {
		ms = append(ms, bson.M{"_id": ""})
	} else {
		condition["$or"] = ms
	}

	// c := bson.M{"$or": []bson.M{{"name": bson.RegEx{Pattern: fmt.Sprintf("zhe1")}}}}

	results, err := d.dao.FindDoc(d.ColName, condition, Page{})
	if err != nil {
		return err
	}
	BsonMapToJson(results)

	return nil
}

// 聚合查询
func (d *UserDao) PipeSearchDemo() error {
	pipes := []bson.M{
		{"$match": bson.M{"age": bson.M{"$gt": 0, "$lt": 8}}},
		{"$group": bson.M{"_id": "$name", "total": bson.M{"$sum": 1}}},
	}
	results, err := d.dao.PipeDoc(d.ColName, pipes)
	if err != nil {
		return err
	}
	BsonMapToJson(results)

	return nil
}

// 按GridFS规范存取文件
func (d *UserDao) GridFsDemo() error {
	id, err := d.dao.CreateGridFs("file.txt", []byte("你住的巷子里，我租了一间公寓"))
	bt, err := d.dao.FindGridFs(id)
	println(string(bt))

	data, err := ioutil.ReadFile(`D:\setup\Caddy\caddy.exe`)
	_, err = d.dao.CreateGridFs("caddy.exe", data)

	return err
}

// TestMgoError 测试mgo数据库查询时，返回的错误
// FindOne 没有匹配到结果时，error 返回 'not found'
// FindDoc 没有匹配到结果时，
func (d *UserDao) TestMgoLibError() error {
	result, err := d.dao.FindOneDoc(d.ColName, bson.ObjectIdHex("5a73c9abc7f41c3744443339"))
	fmt.Printf("errors: %v\n", err)     // not found
	fmt.Printf("result: %+v\n", result) // nil

	result, err = d.dao.FindOneDoc(d.ColName, bson.M{"_id": bson.ObjectIdHex("5a73c9abc7f41c3744443339"), "name": "xx"})
	fmt.Printf("errors: %v\n", err)     // not found
	fmt.Printf("result: %+v\n", result) // nil

	result, err = d.dao.FindDoc(d.ColName, bson.M{}, Page{})
	fmt.Printf("errors: %v\n", err)     // nil
	fmt.Printf("result: %+v\n", result) // [map[xxx][xxx]]

	result, err = d.dao.FindDoc(d.ColName, bson.M{"name": "xxx"}, Page{})
	fmt.Printf("errors: %v\n", err)     // nil
	fmt.Printf("result: %+v\n", result) // []

	result, err = d.dao.FindDoc(d.ColName, bson.M{"_id": "5a73c9abc7f41c3744443339"}, Page{})
	fmt.Printf("errors: %v\n", err)     // nil
	fmt.Printf("result: %+v\n", result) // []

	return nil
}

// TestMgoError 测试mgo数据库查询时，返回的错误
//
// Summary:
// FindId(Or Find(x).One(x) 即结果只有一个) & Update & Delete 时：
//      正常情况：如果查询条件未匹配到结果，err 都会返回 'not found' & result 返回 'nil'
// Find(x).All(&slice)时：
//      正常情况：如果查询条件为匹配到结果，err 都会返回 'nil' & result 返回 '[]'
func (d *UserDao) TestMgoError() error {
	session := d.dao.SessionCopy()
	defer session.Close()

	var i interface{}
	var err error
	col := d.dao.GetCollection("users", session)

	err = col.Insert(1)
	fmt.Printf("errors: %v\n\n", err) // error parsing element 0 of field documents :: caused by :: wrong type
	// for '0' field, expected object, found 0: 1

	var r = model.User{}
	err = col.FindId("5ad4030fc7f41c2920eeccd4").One(&r)
	fmt.Printf("errors: %v\n", err)  // not found
	fmt.Printf("result: %+v\n\n", i) // nil

	i = nil
	err = col.FindId(bson.ObjectIdHex("5ad4030fc7f41c2920eeccd4")).One(&i)
	fmt.Printf("errors: %v\n", err)  // nil
	fmt.Printf("result: %+v\n\n", i) // map[xxx]xxx

	var s []interface{}
	err = col.Find(bson.M{"nam": ""}).All(&s)
	fmt.Printf("errors: %v\n", err)  // nil
	fmt.Printf("result: %+v\n\n", s) // []

	var ss []interface{}
	err = col.Find(bson.M{}).Limit(2).All(&ss)
	fmt.Printf("errors: %v\n", err)   // nil
	fmt.Printf("result: %+v\n\n", ss) // [map[xxx]xxx]

	err = col.UpdateId("5ad4030fc7f41c2920eeccd4", bson.M{"you": "you"})
	fmt.Printf("errors: %v\n\n", err) // not found

	return nil
}

// Response 数据库查询结果处理, 实现 Marshaler
//
// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Response struct {
	Total int         `json:"total"`
	Data  interface{} `json:"data"`
}

// MarshalJSON
func (r Response) MarshalJSON() ([]byte, error) {
	kind := reflect.TypeOf(r.Data).Kind()

	if kind == reflect.Map {
		result := r.Data.(bson.M)
		do(result)
		r.Data = []bson.M{result}
	}

	if kind == reflect.Slice {
		results, ok := r.Data.([]bson.M)
		if ok {
			for _, v := range results {
				do(v)
			}
			r.Data = results
		}
	}

	return json.Marshal(&struct {
		Total int         `json:"total"`
		Data  interface{} `json:"data"`
	}{
		Total: r.Total,
		Data:  r.Data,
	})
}

func do(m bson.M) {
	for key, value := range m {
		if key == "_id" {
			m["id"] = value
			delete(m, "_id")
			break
		}
	}
}

// TestFindOneResultJsonMarshal 数据库查找结果进行Json序列化
func (d *UserDao) TestFindOneResultJsonMarshal() error {
	result, err := d.dao.FindOneDoc(d.ColName, bson.M{"account": "mongo_1"})
	if err != nil {
		return err
	}

	resp := Response{Total: 1, Data: result}
	data, err := resp.MarshalJSON()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(data))

	return err
}

func (d *UserDao) TestFindAllResultJsonMarshal() error {
	result, err := d.dao.FindDoc(d.ColName, bson.M{}, Page{})
	if err != nil {
		return err
	}
	results := result.([]bson.M)

	resp := Response{Total: len(results), Data: result}
	data, err := resp.MarshalJSON()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(data))

	return nil
}
