package models

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type YuqingAnalyse struct {
	Id         string `bson:"_id" form:"_id" json:"_id"` // 主键 词语
	Type       []int  `bson:"t" form:"t" json:"t"`       // 类型 -1断句 0形容词 1名词 2动词 3副词 4否定词 5介词 6助词
	IgnoreType []int  `bson:"i" form:"i" json:"i"`       // 屏蔽情感倾向的词性
	Power      int    `bson:"p" form:"p" json:"p"`       // 程度
}

// 查询全部
func (this *YuqingAnalyse) All() []*YuqingAnalyse {
	var r []*YuqingAnalyse
	Db.C("yuqing.analyse").Find(nil).All(&r)

	return r
}

////////////////////////////////////////////////////////////////////////////////
// restful api
////////////////////////////////////////////////////////////////////////////////

// 解析数组
func (this *YuqingAnalyse) BaseFormStrings(contro *beego.Controller) (bool, interface{}) {
	types := contro.GetStrings("t")

	if len(types) > 0 {
		typeInts := make([]int, len(types))

		for k, _ := range types {
			typeInts[k], _ = strconv.Atoi(types[k])
		}

		this.Type = typeInts
	}

	ignoreTypes := contro.GetStrings("i")

	if len(ignoreTypes) > 0 {
		typeInts := make([]int, len(ignoreTypes))

		for k, _ := range ignoreTypes {
			typeInts[k], _ = strconv.Atoi(ignoreTypes[k])
		}

		this.IgnoreType = typeInts
	}

	return true, nil
}

func (this *YuqingAnalyse) BaseCount() int {
	count, err := Db.C("yuqing.analyse").Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *YuqingAnalyse) BaseFilterCount(filterKey string, filter interface{}) int {
	count, err := Db.C("yuqing.analyse").Find(bson.M{filterKey: filter}).Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *YuqingAnalyse) BaseInsert() error {
	err := Db.C("yuqing.analyse").Insert(this)
	return err
}

func (this *YuqingAnalyse) BaseFind(id string) (bool, interface{}) {

	err := Db.C("yuqing.analyse").FindId(id).One(&this)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (this *YuqingAnalyse) BaseUpdate() error {
	err := Db.C("yuqing.analyse").UpdateId(this.Id, this)
	return err
}

func (this *YuqingAnalyse) BaseDelete() error {
	err := Db.C("yuqing.analyse").RemoveId(this.Id)
	return err
}

/* 设置objectid */
func (this *YuqingAnalyse) BaseSetId(id string) (bool, interface{}) {
	this.Id = id
	return true, nil
}

func (this *YuqingAnalyse) BaseSelect(from int, limit int, sort string) []BaseInterface {
	var r []*YuqingAnalyse
	err := Db.C("yuqing.analyse").Find(nil).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingAnalyse) BaseFilterSelect(from int, limit int, sort string, filterKey string, filter interface{}) []BaseInterface {
	var r []*YuqingAnalyse
	err := Db.C("yuqing.analyse").Find(bson.M{filterKey: filter}).Sort(sort).Skip(from).Limit(limit).All(&r)
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
func (this *YuqingAnalyse) BaseSelectLike(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*YuqingAnalyse

	err := Db.C("yuqing.analyse").Find(bson.M{target: bson.M{"$regex": key}}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingAnalyse) BaseLikeCount(target string, key interface{}) int {
	count, err := Db.C("yuqing.analyse").Find(bson.M{target: bson.M{"$regex": key}}).Count()
	if err != nil {
		return 0
	}
	return count
}

/* 精确匹配 */
func (this *YuqingAnalyse) BaseSelectAccuracy(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*YuqingAnalyse

	err := Db.C("yuqing.analyse").Find(bson.M{target: key}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *YuqingAnalyse) BaseAccuracyCount(target string, key interface{}) int {
	count, err := Db.C("yuqing.analyse").Find(bson.M{target: key}).Count()
	if err != nil {
		return 0
	}
	return count
}
