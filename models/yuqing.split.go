package models

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type YuqingSplit struct {
	Id   string `bson:"_id" json:"_id" form:"_id"` // 主键 分词
	Hot  int    `bson:"h" json:"h" form:"h"`       // 频率
	Type string `bson:"t" json:"t" form:"t"`       // 词性
}

// 查询全部
func (this *YuqingSplit) All() []*YuqingSplit {
	var r []*YuqingSplit
	Db.C("yuqing.split").Find(nil).All(&r)

	return r
}

// 实现载入字典接口
func (this *YuqingSplit) GetText() string {
	return this.Id
}

func (this *YuqingSplit) GetFrequency() int {
	return this.Hot
}

func (this *YuqingSplit) GetPos() string {
	return this.Type
}

////////////////////////////////////////////////////////////////////////////////
// restful api
////////////////////////////////////////////////////////////////////////////////

// 解析数组
func (this *YuqingSplit) BaseFormStrings(contro *beego.Controller) (bool, interface{}) {
	return true, nil
}

func (this *YuqingSplit) BaseCount() int {
	count, err := Db.C("yuqing.split").Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *YuqingSplit) BaseFilterCount(filterKey string, filter interface{}) int {
	count, err := Db.C("yuqing.split").Find(bson.M{filterKey: filter}).Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *YuqingSplit) BaseInsert() error {
	err := Db.C("yuqing.split").Insert(this)
	return err
}

func (this *YuqingSplit) BaseFind(id string) (bool, interface{}) {

	err := Db.C("yuqing.split").FindId(id).One(&this)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (this *YuqingSplit) BaseUpdate() error {
	err := Db.C("yuqing.split").UpdateId(this.Id, this)
	return err
}

func (this *YuqingSplit) BaseDelete() error {
	err := Db.C("yuqing.split").RemoveId(this.Id)
	return err
}

/* 设置objectid */
func (this *YuqingSplit) BaseSetId(id string) (bool, interface{}) {
	this.Id = id
	return true, nil
}

func (this *YuqingSplit) BaseSelect(from int, limit int, sort string) []BaseInterface {
	var r []*YuqingSplit
	err := Db.C("yuqing.split").Find(nil).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingSplit) BaseFilterSelect(from int, limit int, sort string, filterKey string, filter interface{}) []BaseInterface {
	var r []*YuqingSplit
	err := Db.C("yuqing.split").Find(bson.M{filterKey: filter}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

/* 模糊匹配 */
func (this *YuqingSplit) BaseSelectLike(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*YuqingSplit
	err := Db.C("yuqing.split").Find(bson.M{target: bson.M{"$regex": key}}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingSplit) BaseLikeCount(target string, key interface{}) int {
	count, err := Db.C("yuqing.split").Find(bson.M{target: bson.M{"$regex": key}}).Count()
	if err != nil {
		return 0
	}
	return count
}

/* 精确匹配 */
func (this *YuqingSplit) BaseSelectAccuracy(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*YuqingSplit

	err := Db.C("yuqing.split").Find(bson.M{target: key}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingSplit) BaseAccuracyCount(target string, key interface{}) int {
	count, err := Db.C("yuqing.split").Find(bson.M{target: key}).Count()
	if err != nil {
		return 0
	}
	return count
}
