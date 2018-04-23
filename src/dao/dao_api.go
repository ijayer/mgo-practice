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
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gedex/inflector"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Dao Database Access Object
type Dao struct {
	Name     string       // 数据库名称
	Session  *mgo.Session // 数据库连接池
	PrefixFS string       // GridFS前缀
}

// NewDao 初始化Dao对象
func NewDao(session *mgo.Session) *Dao {
	return &Dao{
		Session:  session,
		Name:     DBCfg.Name,
		PrefixFS: fmt.Sprintf("fs"),
	}
}

// SessionCopy 从源Session完成拷贝(该拷贝保留原有Session信息)
func (d *Dao) SessionCopy() *mgo.Session {
	return d.Session.Copy()
}

// GetDB 获取mgo.Database对象
func (d *Dao) GetDB(session *mgo.Session) *mgo.Database {
	return d.Session.DB(d.Name)
}

// DropDB 删除数据库
func (d *Dao) DropDB() error {
	return d.Session.DB(d.Name).DropDatabase()
}

// GetCollection 获取mgo.Collection对象
func (d *Dao) GetCollection(name string, session *mgo.Session) *mgo.Collection {
	if name == "" {
		name = fmt.Sprint("mongos")
	}
	return d.GetDB(session).C(name)
}

var (
	errNull          = errors.New("the interface is nil")
	errUnSupportType = errors.New("unsupported type(only support bson.ObjectId or bson.M)")
)

/*
 * 封装 mgo 相关函数
 */

// ResultWithMap 数据库查询结果存入map[string]interface{}
type ResultWithMap map[string]interface{}

// CreateDoc 插入文档
// name 集合名；docs 要插入的文档；keys 索引字段
//
// TODO: [20180423]数据库有关时间的字段应该存储为：时间戳，然后在代码中封装时间格式转换函数
func (d *Dao) CreateDoc(collection string, docs interface{}, keys ...string) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(collection)

	if len(keys) == 0 {
		keys = append(keys, "-create_at")
		// Warn: "-create_at["2006-01-02 15:04:05"]" maybe caused duplicated index keys
	}
	index := mgo.Index{
		Key:        keys, // 索引键
		Unique:     true, // 创建唯一索引
		DropDups:   true, // 删除重复索引
		Background: true, // 在后台创建
		Sparse:     true, // 不存在字段不启用索引
	}
	if err := co.EnsureIndex(index); err != nil {
		return err
	}

	return co.Insert(docs)
}

// UpsertDoc 插入 & 更新文档
// name 指定集合名；selector 选择条件；update 更新内容
//
// UpsertDoc 方法内部采用了两种实现方式
// 1：调用 session.DB(name).C(collection).Upsert 方法
// 2：调用 session.DB(name).C(collection).Find(selector).Apply() 方法
//    Apply()方法底层实际运行了`findAndModify`命令：
func (d *Dao) UpsertDoc(name string, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if selector == nil {
		return nil, errNull
	}
	if m, ok := selector.(bson.M); ok { // selector 为 bson.M
		if change, ok := update.(mgo.Change); ok {
			// 支持 mgo.Change
			// $setOnInsert 设置只在文档创建时需要添加的字段
			// change := mgo.Change{
			// 		Update: bson.M{
			// 			"$set":         update,
			// 			"$setOnInsert": bson.M{"create_at": Now(), "is_delete": false, "delete_at": ""},
			// 		},
			// 		Upsert:    true,
			// 		ReturnNew: true,
			// 	}
			var i interface{}
			return co.Find(m).Apply(change, &i)
		}
		return co.Upsert(m, update)
	}
	if id, ok := selector.(bson.ObjectId); ok { // selector 为 bson.ObjectId
		if change, ok := update.(mgo.Change); ok {
			var i interface{}
			return co.FindId(id).Apply(change, &i)
		}
		return co.UpsertId(id, update)
	}
	return nil, errUnSupportType
}

// RemoveDoc 删除文档，物理删除
// name 集合名；selector 选择条件(selector 存储 bson.ObjectId or bson.M 类型)
func (d *Dao) RemoveDoc(name string, selector interface{}) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

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

// RemoveDocByMark 软删除文档
// name 集合名；selector 选择条件(selector 存储 bson.ObjectId or bson.M 类型)
func (d *Dao) RemoveDocByMark(name string, selector interface{}) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

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

// UpdateDoc 更新文档
// name 集合名；selector 选择条件(selector 存储 bson.ObjectId or bson.M 类型); update 更新内容
func (d *Dao) UpdateDoc(name string, selector interface{}, update interface{}) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if selector == nil || update == nil {
		return errNull
	}

	docs, ok := update.(bson.M)
	if ok {
		if _, ok := docs["_id"]; ok {
			delete(docs, "_id")
		}
		if _, ok := docs["create_at"]; ok {
			delete(docs, "create_at")
		}
		update = docs
	}

	if m, ok := selector.(bson.M); ok {
		return co.Update(m, bson.M{"$set": update})
	}
	if id, ok := selector.(bson.ObjectId); ok {
		return co.UpdateId(id, bson.M{"$set": update})
	}
	return errUnSupportType
}

// Page 定义分页查询参数存储对象
type Page struct {
	Valid         bool
	Offset, Limit int
}

// checkValid 检查分页参数是否合法
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

// FindWithQuery 查询文档，其结果存入mgo.Query返回
// name集合名称; query查询条件；page分页条件；sortKeys排序字段。该方法将返回按条件过滤后的 *mgo.Query 结构
func (d *Dao) FindWithQuery(name string, query interface{}, page Page, sortKeys ...string) (*mgo.Query, error) {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if query == nil {
		return nil, errNull
	}
	q := co.Find(query)

	if len(sortKeys) == 0 {
		sortKeys = append(sortKeys, "-create_at")
	}
	q = q.Sort(sortKeys...)

	if page.Valid {
		q = q.Skip(page.Offset).Limit(page.Limit)
	}
	return q, nil
}

// FindDoc 查找文档, 其结果存入[]interface返回
// name集合名称; query查询条件; page指定分页参数; sortKeys指定排序字段
func (d *Dao) FindDoc(name string, query interface{}, page Page, sortKeys ...string) (interface{}, error) {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if query == nil {
		return nil, errNull
	}
	q := co.Find(query)

	if len(sortKeys) == 0 {
		sortKeys = append(sortKeys, "-create_at")
	}
	q = q.Sort(sortKeys...)

	var err error
	var results []bson.M
	if page.Valid {
		q = q.Skip(page.Offset).Limit(page.Limit)
	}
	err = q.All(&results)

	return results, err
}

// FindDocToResults 查找文档，其结果写入result(结构体对象的指针的切片)，并返回一个error
// name集合名称; query查询条件; page指定分页参数; sortKeys指定排序字段
func (d *Dao) FindDocToResults(name string, results, query interface{}, page Page, sortKeys ...string) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if reflect.TypeOf(results).Kind() != reflect.Ptr {
		return fmt.Errorf("results must be a pointer")
	}

	if query == nil {
		return errNull
	}
	q := co.Find(query)

	if len(sortKeys) == 0 {
		sortKeys = append(sortKeys, "-create_at")
	}
	q = q.Sort(sortKeys...)

	if page.Valid {
		q = q.Skip(page.Offset).Limit(page.Limit)
	}
	return q.All(results)
}

// FindOneDoc 查找某个文档, interface{}存储的结果为bson.M格式
// name集合名称; query指定查询条件(contains _id or an unique_main_key)
func (d *Dao) FindOneDoc(name string, query interface{}) (interface{}, error) {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if query == nil {
		return nil, errNull
	}

	var err error
	var q *mgo.Query
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

	var result bson.M
	if id, ok := query.(bson.ObjectId); ok {
		q = co.FindId(id)
	}
	err = q.One(&result)

	return result, err
}

// FindOneDocToResult 查找某个文档, 其结果写入result(结构体对象的指针)，并返回一个error
// name集合名称; query指定查询条件(contains _id or an unique_main_key)
func (d *Dao) FindOneDocToResult(name string, result, query interface{}) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("results must be a pointer type")
	}

	if query == nil {
		return errNull
	}

	var q *mgo.Query
	if m, ok := query.(bson.M); ok {
		q = co.Find(m)
		cnt, err := q.Count()
		if err != nil {
			return err
		}
		if cnt > 1 {
			return mgo.ErrNotFound
		}
	}

	if id, ok := query.(bson.ObjectId); ok {
		q = co.FindId(id)
	}
	return q.One(result)
}

// PipeDoc 聚合管道
// name集合名称; pipes指定管道操作条件
func (d *Dao) PipeDoc(name string, pipes []bson.M) (interface{}, error) {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	var err error
	var results []bson.M

	err = co.Pipe(pipes).All(&results)
	return results, err
}

// PipeOneDocToResult 聚合管道, 其结果写入result(结构体对象的指针)，并返回一个error
// name集合名称; pipes指定管道操作条件
func (d *Dao) PipeOneDocToResult(name string, pipes []bson.M, result interface{}) error {
	session := d.SessionCopy()
	defer session.Close()
	co := session.DB(d.Name).C(name)

	if reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("results must be a pointer type")
	}
	return co.Pipe(pipes).One(result)
}

// CreateGridFs 存储文件 GridFS
// name 文件名; writer o.ReadWriter接口; 返回文档 Id 和 error
func (d *Dao) CreateGridFs(name string, data []byte) (bson.ObjectId, error) {
	session := d.SessionCopy()
	defer session.Close()
	gfs := session.DB(d.Name).GridFS(d.PrefixFS)

	id := bson.NewObjectId()
	fs, err := gfs.Create(name)
	if err != nil {
		return id, err
	}
	fs.SetId(id)

	_, err = fs.Write(data)
	if err != nil {
		return id, err
	}
	if err := fs.Close(); err != nil {
		return "", err
	}
	return id, nil
}

// FindGridFs 查找文件，文档id
func (d *Dao) FindGridFs(id interface{}) ([]byte, error) {
	session := d.SessionCopy()
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

// DBRef 解析mgo.DBRef，其结果写入 m(map[string]interface{})
func DBRef(field string, t reflect.Type, m map[string]interface{}) error {
	if value, hit := m[field]; hit {
		refer, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid param given [%s]", field)
		}
		id, hit := refer["id"]
		if !hit {
			return fmt.Errorf("%s must be an object and contain a id field", field)
		}
		delete(m, field)

		if reflect.TypeOf(id).Kind() != reflect.String || !bson.IsObjectIdHex(id.(string)) {
			return fmt.Errorf("id format error [%v]", id)
		}
		m[field+"_ref"] = mgo.DBRef{
			Id:         bson.ObjectIdHex(id.(string)),
			Collection: strings.ToLower(inflector.Pluralize(t.Name())),
		}
	}
	return nil
}

// DBRefId 解析 mgo.DBRef.Id (mgo_key: $id)
func DBRefId(field string, m map[string]interface{}) error {
	if value, hit := m[field]; hit {
		refer, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid param given [%s]", field)
		}
		id, hit := refer["id"]
		if !hit {
			return fmt.Errorf("%s must be an object and contain a id field", field)
		}
		delete(m, field)

		if reflect.TypeOf(id).Kind() != reflect.String || !bson.IsObjectIdHex(id.(string)) {
			return fmt.Errorf("id format error [%v]", id)
		}
		m[field+"_ref.$id"] = bson.ObjectIdHex(id.(string))
	}
	return nil
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
