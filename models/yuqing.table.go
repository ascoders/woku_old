package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type YuqingTable struct {
	Id       bson.ObjectId `bson:"_id" json:"_id" form:"_id"` // 主键
	Category string        `bson:"c" json:"c" form:"c"`       // 所属分类
	Good     int           `bson:"g" json:"g" form:"g"`       // 积极数量
	Normal   int           `bson:"n" json:"n" form:"n"`       // 中立数量
	Bad      int           `bson:"b" json:"b" form:"b"`       // 消极数量
	Created  time.Time     `bson:"cr" json:"cr" form:"cr"`    // 创建日期
}

// 查询全部
func (this *YuqingTable) All() []*YuqingTable {
	var r []*YuqingTable
	Db.C("yuqing.table").Find(nil).All(&r)

	return r
}

// 插入
func (this *YuqingTable) Insert() bool {
	this.Id = bson.NewObjectId()
	this.Created = time.Now()

	err := Db.C("yuqing.table").Insert(this)
	return err == nil
}

// 删除一个星期之外的信息
func (this *YuqingTable) RemoveOld() bool {
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)

	err, _ := Db.C("yuqing.table").RemoveAll(bson.M{"cr": bson.M{"$lt": weekAgo}})
	return err == nil
}

// 查询某个分类一星期内信息
func (this *YuqingTable) Find(category string) []*YuqingTable {
	var result []*YuqingTable

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)

	Db.C("yuqing.table").Find(bson.M{"c": category, "cr": bson.M{"$gt": weekAgo}}).Sort("cr").All(&result)

	return result
}

// 查询某分类最近一次信息
func (this *YuqingTable) FindNew(category string) *YuqingTable {
	var result *YuqingTable

	Db.C("yuqing.table").Find(bson.M{"c": category}).Sort("cr").One(&result)

	return result
}
