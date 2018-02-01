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
type MongoConfig struct {
	DBName       string // 数据库名称
	DBAdds       addrs  // 数据库地址
	DBUsername   string // 数据库账号
	DBPassword   string // 数据库密码
	EnableAuth   bool   // 是否启用数据库验证
	RepSetName   string // 副本集(Replica set)名称
	EnableRepSet bool   // 是否启用Replica Set集群模式
}

// DBConfig 表示一个MongoDB的全局配置对象
var DBConfig = &MongoConfig{}

// 初始化MongoDB配置信息
func initMongoConfig() {
	flag.Var(&DBConfig.DBAdds, "db_addr", "database cluster server address")
	flag.BoolVar(&DBConfig.EnableAuth, "db_auth", false, "enable database authorization")
	flag.BoolVar(&DBConfig.EnableRepSet, "db_rs", false, "enable replica set")
	flag.StringVar(&DBConfig.DBName, "db_name", "mongo", "database name for your app")
	flag.StringVar(&DBConfig.DBUsername, "username", "mongo", "database username")
	flag.StringVar(&DBConfig.DBPassword, "password", "mongo", "database password ")
	flag.StringVar(&DBConfig.RepSetName, "rs", "rs", "replica set name")
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
	info.Database = DBConfig.DBName
	if len(DBConfig.DBAdds) == 0 {
		DBConfig.DBAdds = append(DBConfig.DBAdds, "127.0.0.1:27017")
	}
	info.Addrs = DBConfig.DBAdds

	// Enable database authorization
	if DBConfig.EnableAuth {
		info.Username = DBConfig.DBUsername
		info.Password = DBConfig.DBPassword
	}

	// Enable replica set
	if DBConfig.EnableRepSet {
		info.ReplicaSetName = DBConfig.RepSetName
	}

	session, err = mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic(单调的) behavior(行为).
	session.SetMode(mgo.Monotonic, true)

	return session
}
