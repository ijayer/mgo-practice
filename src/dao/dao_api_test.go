/*
 * 说明：数据库公用接口单元测试
 * 作者：zhe
 * 时间：2018-01-30 13:47
 * 更新：
 */

package dao

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestMain(m *testing.M) {
	session := InitMongo()
	defer session.Close()

	clean(session)

	code := m.Run()
	os.Exit(code)
}

func clean(session *mgo.Session) {
	err := session.DB("mongo").DropDatabase()
	if err != nil {
		panic(err)
	}
}

func TestNewDao(t *testing.T) {
	session := InitMongo()
	defer session.Close()

	type args struct {
		session *mgo.Session
	}
	tests := []struct {
		name string
		args args
		want *Dao
	}{
		{
			name: "NewDao",
			args: struct{ session *mgo.Session }{session: session},
			want: NewDao(session),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDao(tt.args.session); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDao() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDao_CreateDoc(t *testing.T) {
	session := InitMongo()
	defer session.Close()

	type fields struct {
		Name    string
		Session *mgo.Session
	}
	type args struct {
		collection string
		docs       interface{}
		idxKeys    []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "CreateDoc",
			fields: struct {
				Name    string
				Session *mgo.Session
			}{Name: "mongo", Session: session},
			args: struct {
				collection string
				docs       interface{}
				idxKeys    []string
			}{collection: "mongos", docs: bson.M{"first": "a", "second": "b"}, idxKeys: []string{"first"}},
		},
		{
			name: "CreateDoc",
			fields: struct {
				Name    string
				Session *mgo.Session
			}{Name: "mongo", Session: session},
			args: struct {
				collection string
				docs       interface{}
				idxKeys    []string
			}{collection: "mongos", docs: bson.M{"first": "c", "second": "d"}, idxKeys: []string{"first"}},
		},
		{
			name: "CreateDoc",
			fields: struct {
				Name    string
				Session *mgo.Session
			}{Name: "mongo", Session: session},
			args: struct {
				collection string
				docs       interface{}
				idxKeys    []string
			}{collection: "mongos", docs: bson.M{"one": 1, "two": 2}, idxKeys: []string{"one"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				Name:    tt.fields.Name,
				Session: tt.fields.Session,
			}
			if err := d.CreateDoc(tt.args.collection, tt.args.docs, tt.args.idxKeys...); (err != nil) != tt.wantErr {
				t.Errorf("Dao.CreateDoc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDao_UpsertDoc(t *testing.T) {
	session := InitMongo()
	defer session.Close()

	type fields struct {
		Name    string
		Session *mgo.Session
	}
	fs := fields{
		Name:    "mongo",
		Session: session,
	}

	type args struct {
		collection string
		selector   interface{}
		update     interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *mgo.ChangeInfo
		wantErr bool
	}{
		{
			name:   "UpsertDoc",
			fields: fs,
			args: struct {
				collection string
				selector   interface{}
				update     interface{}
			}{collection: "mongos", selector: bson.M{"first": "a"}, update: bson.M{"$set": bson.M{"second": 2}}},
			wantErr: false,
		},
		{
			name:   "UpsertDoc",
			fields: fs,
			args: struct {
				collection string
				selector   interface{}
				update     interface{}
			}{collection: "mongos", selector: "11223344", update: bson.M{"second": 2}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				Name:    tt.fields.Name,
				Session: tt.fields.Session,
			}
			_, err := d.UpsertDoc(tt.args.collection, tt.args.selector, tt.args.update)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dao.UpsertDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDao_UpdateDoc(t *testing.T) {
	session := InitMongo()
	defer session.Close()

	type fields struct {
		Name    string
		Session *mgo.Session
	}
	type args struct {
		collection string
		selector   interface{}
		update     bson.M
	}
	fs := fields{
		Name:    "mongo",
		Session: session,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			fields:  fs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				Name:    tt.fields.Name,
				Session: tt.fields.Session,
			}
			if err := d.UpdateDoc(tt.args.collection, tt.args.selector, tt.args.update); (err != nil) != tt.wantErr {
				t.Errorf("Dao.UpdateDoc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
