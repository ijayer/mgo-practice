/*
 * 说明：数据访问对象（Data Access Object，DAO）
 * 作者：zhe
 * 时间：2018-01-17 23:10
 * 更新：数据库连接&初始化
 */

package dao

import (
	"flag"
	"time"

	"gopkg.in/mgo.v2"
)

// 定义MongoDB的配置信息
type MongoDB struct {
	Name         string // 数据库名称
	Adds         addrs  // 数据库地址
	Username     string // 数据库账号
	Password     string // 数据库密码
	EnableAuth   bool   // 是否启用数据库验证
	RepSetName   string // 副本集(Replica set)名称
	EnableRepSet bool   // 是否启用Replica Set集群模式
}

// DBConfig 表示一个MongoDB的全局配置对象
var DBCfg = &MongoDB{}

// 初始化MongoDB配置信息
func initMongoConfig() {
	flag.Var(&DBCfg.Adds, "db_addr", "database cluster server address")
	flag.BoolVar(&DBCfg.EnableAuth, "db_auth", false, "enable database authorization")
	flag.BoolVar(&DBCfg.EnableRepSet, "db_rs", false, "enable replica set")
	flag.StringVar(&DBCfg.Name, "db_name", "mongo", "database name for your app")
	flag.StringVar(&DBCfg.Username, "username", "mongo", "database username")
	flag.StringVar(&DBCfg.Password, "password", "mongo", "database password ")
	flag.StringVar(&DBCfg.RepSetName, "rs", "rs", "replica set name")
	flag.Parse()
}

func init() {
	initMongoConfig()
}

// InitMongo 根据命令行参数初始化数据库连接，失败则引发panic, 成功返回mgo连接池(session)
func InitMongo() *mgo.Session {
	var err error
	var session *mgo.Session

	info := &mgo.DialInfo{}

	// Dial with sever timeout
	info.Timeout = 30 * time.Second

	// Init database cluster server address
	info.Database = DBCfg.Name
	if len(DBCfg.Adds) == 0 {
		DBCfg.Adds = append(DBCfg.Adds, "127.0.0.1:27017")
	}
	info.Addrs = DBCfg.Adds

	// Enable database authorization
	if DBCfg.EnableAuth {
		info.Username = DBCfg.Username
		info.Password = DBCfg.Password
	}

	// Enable replica set
	if DBCfg.EnableRepSet {
		info.ReplicaSetName = DBCfg.RepSetName
	}

	session, err = mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic(单调的) behavior(行为).
	session.SetMode(mgo.Monotonic, true)

	return session
}
