/*
 * 说明：数据库公共接口实现
 * 作者：zhe
 * 时间：2018-01-17 23:16
 * 更新：
 */

package dao

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Database Access Object
type Dao struct {
	Name     string       // 数据库名称
	Session  *mgo.Session // 数据库连接池
	PrefixFS string       // GridFS前缀
}

// 初始化Dao对象
func NewDao(session *mgo.Session) *Dao {
	return &Dao{
		Session:  session,
		Name:     DBConfig.DBName,
		PrefixFS: fmt.Sprintf("fs"),
	}
}

// 从源Session完成拷贝(该拷贝保留原有Session信息)
func (d *Dao) sessionCopy() *mgo.Session {
	return d.Session.Copy()
}

// 获取mgo.Database对象
func (d *Dao) getDB() *mgo.Database {
	return d.Session.DB(d.Name)
}

// 删除数据库
func (d *Dao) DropDB() error {
	return d.Session.DB(d.Name).DropDatabase()
}

// 获取mgo.Collection对象
func (d *Dao) getCollection(name string) *mgo.Collection {
	if name == "" {
		name = fmt.Sprint("mongos")
	}
	return d.Session.DB(d.Name).C(name)
}

var (
	errNull          = errors.New("the interface is nil")
	errUnSupportType = errors.New("unsupported type(only support bson.ObjectId or bson.M)")
)

/*
 * 封装 mgo 相关函数
 */

// 插入文档
func (d *Dao) CreateDoc(collection string, docs interface{}, idxKeys ...string) error {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if len(idxKeys) == 0 {
		idxKeys = append(idxKeys, "-create_at")
	}
	index := mgo.Index{
		Key:        idxKeys, // 索引键
		Unique:     true,    // 创建唯一索引
		DropDups:   true,    // 删除重复索引
		Background: true,    // 在后台创建
		Sparse:     true,    // 不存在字段不启用索引
	}
	if err := co.EnsureIndex(index); err != nil {
		return err
	}

	return co.Insert(docs)
}

// 插入 & 更新文档
// Method1：调用 session.DB(name).C(collection).Upsert 方法
// Method2：调用 session.DB(name).C(collection).Find(selector).Apply() 方法
//          Apply()方法底层实际运行了`findAndModify`命令：
func (d *Dao) UpsertDoc(collection string, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if selector == nil {
		return nil, errNull
	}
	if m, ok := selector.(bson.M); ok {
		if change, ok := update.(mgo.Change); ok {
			var i interface{}
			return co.Find(m).Apply(change, &i)
		}
		return co.Upsert(m, update)
	}
	if id, ok := selector.(bson.ObjectId); ok {
		if change, ok := update.(mgo.Change); ok {
			var i interface{}
			return co.FindId(id).Apply(change, &i)
		}
		return co.UpsertId(id, update)
	}
	return nil, errUnSupportType
}

// 删除文档, i 存储 bson.ObjectId or bson.M 类型
func (d *Dao) RemoveDoc(collection string, selector interface{}) error {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if selector == nil {
		return errNull
	}
	if m, ok := selector.(bson.M); ok {
		return co.Remove(m)
	}
	if id, ok := selector.(bson.ObjectId); ok {
		return co.RemoveId(id)
	}
	return errUnSupportType
}

// 软删除文档
func (d *Dao) SoftRemoveDoc(collection string, selector interface{}) error {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if selector == nil {
		return errNull
	}

	update := bson.M{}
	update["modify_at"] = Now()
	update["delete_at"] = Now()
	update["is_delete"] = true
	if m, ok := selector.(bson.M); ok {
		return co.Update(m, bson.M{"$set": update})
	}
	if id, ok := selector.(bson.ObjectId); ok {
		return co.UpdateId(id, bson.M{"$set": update})
	}
	return errUnSupportType
}

// 更新文档
func (d *Dao) UpdateDoc(collection string, selector interface{}, update bson.M) error {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if selector == nil {
		return errNull
	}
	if _, ok := update["_id"]; ok {
		delete(update, "_id")
	}
	if _, ok := update["create_at"]; ok {
		delete(update, "create_at")
	}

	if m, ok := selector.(bson.M); ok {
		return co.Update(m, update)
	}
	if id, ok := selector.(bson.ObjectId); ok {
		return co.UpdateId(id, update)
	}
	return errUnSupportType
}

// 定义分页查询参数存储对象
type Page struct {
	Valid         bool
	Offset, Limit int
}

// 检查分页参数是否合法
func (p *Page) checkValid(offset, limit string) {
	if offset == "" || limit == "" {
		p.Valid = false
		return
	}

	var err error
	var l, o int64

	l, err = strconv.ParseInt(limit, 10, 32)
	p.Limit = int(l)
	if err != nil || p.Limit < 0 {
		p.Valid = false
		return
	}

	o, err = strconv.ParseInt(offset, 10, 32)
	p.Offset = int(o)
	if err != nil || p.Offset < 0 {
		p.Valid = false
		return
	}

	p.Valid = true
}

// 查找文档：query指定查询条件，page指定分页参数，sortKeys指定排序字段
func (d *Dao) FindDoc(collection string, query interface{}, page Page, sortKeys ...string) ([]interface{}, error) {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if query == nil {
		return nil, errNull
	}
	q := co.Find(query)

	if len(sortKeys) == 0 {
		sortKeys = append(sortKeys, "-create_at")
	}
	q = q.Sort(sortKeys...)

	var err error
	var results []interface{}
	if page.Valid {
		q = q.Skip(page.Offset).Limit(page.Limit)
	}
	q.All(&results)

	return results, err
}

// 查找某个文档：query指定查询条件(contains _id or an unique_main_key)，sortKeys指定排序字段
func (d *Dao) FindOne(collection string, query interface{}) (interface{}, error) {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if query == nil {
		return nil, errNull
	}

	var err error
	var q *mgo.Query
	var result interface{}
	if m, ok := query.(bson.M); ok {
		q = co.Find(m)
		cnt, err := q.Count()
		if err != nil {
			return nil, err
		}
		if cnt > 1 {
			return nil, mgo.ErrNotFound
		}
	}
	if id, ok := query.(bson.ObjectId); ok {
		q = co.FindId(id)
	}
	err = q.One(&result)

	return result, err
}

// 聚合管道: collection指定集合名称; pipes指定管道操作条件; page指定分页参数; addition指定额外条件
func (d *Dao) PipeDoc(collection string, pipes []bson.M) ([]interface{}, error) {
	session := d.sessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	var err error
	var results []interface{}

	err = co.Pipe(pipes).All(&results)
	return results, err
}

// 存储文件：GridFS
func (d *Dao) CreateGridFs(name string, writer io.ReadWriter) (interface{}, error) {
	session := d.sessionCopy()
	defer session.Close()
	gfs := session.DB(d.Name).GridFS(d.PrefixFS)

	id := bson.NewObjectId()
	f, err := gfs.Create(name)
	if err != nil {
		return id, err
	}
	f.SetId(id)

	_, err = io.Copy(f, writer)
	if err != nil {
		return id, err
	}
	return id, nil
}

// 查找文件
func (d *Dao) FindGridFs(id interface{}) ([]byte, error) {
	session := d.sessionCopy()
	defer session.Close()
	gfs := session.DB(d.Name).GridFS(d.PrefixFS)

	fs, err := gfs.OpenId(id)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, fs); err != nil {
		return nil, err
	}
	if err := fs.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*
 * 封装 mgo 操作符
 */

type Operator struct {
	ops sort.StringSlice
}

func NewOperator(ops ...string) {
	operators := sort.StringSlice{}
	operators = append(operators, ops...)
}

// 解析 mgo.DBRef
func (o *Operator) DBRef() {

}

// 解析 mgo.DBRef.Id (mgo_key: $id)
func (o *Operator) DBRefId() {

}
