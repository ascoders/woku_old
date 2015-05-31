package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type YuqingResult struct {
	Id        string             `bson:"_id" json:"_id" form:"_id"` // 主键 标题
	Category  string             `bson:"c" json:"c" form:"c"`       // 分类
	Good      int                `bson:"g" json:"g" form:"g"`       // 综合好坏
	Detail    []YuqingResultWord `bson:"d" json:"d" form:"d"`       // 详情
	WordSlice []string           `bson:"w" json:"w" form:"w"`       // 分词词组
	Created   time.Time          `bson:"cr" json:"cr" form:"cr"`    // 创建时间
}

type YuqingResultWord struct {
	Content     string `bson:"c" json:"c" form:"c"` // 词语
	Good        int    `bson:"g" json:"g" form:"g"` // 好坏度
	Type        int    `bson:"t" json:"t" form:"t"` // 类型 （-1断句/0评价/1情感/2主张/3程度副词/4否定）
	Position    int    `bson:"p" json:"p" form:"p"` // 在分词中位置
	DegreeGroup int    `bson:"d" json:"d" form:"d"` // 否定分组（0不表示组）
}

// 删除一个月之外的信息
func (this *YuqingResult) RemoveOld() bool {
	weekAgo := time.Now().Add(-30 * 24 * time.Hour)

	err, _ := Db.C("yuqing.result").RemoveAll(bson.M{"cr": bson.M{"$lt": weekAgo}})
	return err == nil
}

// 插入新纪录
func (this *YuqingResult) Insert() bool {
	// 试图删除旧记录
	Db.C("yuqing.result").Remove(bson.M{"_id": this.Id})

	this.Created = time.Now()

	err := Db.C("yuqing.result").Insert(this)
	return err == nil
}

// 查询某个分类信息
func (this *YuqingResult) Find(category string, from int, number int) []*YuqingResult {
	var result []*YuqingResult

	if category != "" {
		Db.C("yuqing.result").Find(bson.M{"c": category}).Sort("-cr").Skip(from).Limit(number).All(&result)
	} else {
		Db.C("yuqing.result").Find(bson.M{"c": bson.M{"$ne": "test"}}).Sort("-cr").Skip(from).Limit(number).All(&result)
	}

	return result
}

// 查询一天内的记录
func (this *YuqingResult) FindToday() []*YuqingResult {
	var result []*YuqingResult

	dayAgo := time.Now().Add(-1 * 24 * time.Hour)
	Db.C("yuqing.result").Find(bson.M{"cr": bson.M{"$gt": dayAgo}}).All(&result)

	return result
}

// 查询某分类一天内的记录
func (this *YuqingResult) FindCategoryToday(category string) []*YuqingResult {
	var result []*YuqingResult

	dayAgo := time.Now().Add(-1 * 24 * time.Hour)
	Db.C("yuqing.result").Find(bson.M{"c": category, "cr": bson.M{"$gt": dayAgo}}).All(&result)

	return result
}

// 更新
func (this *YuqingResult) Update() {
	Db.C("yuqing.result").Update(bson.M{"_id": this.Id}, this)
}
