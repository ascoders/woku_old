package models

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type Job struct {
	Id          uint8  `bson:"_id" json:"_id" form:"_id"` //主键
	Name        string `bson:"n" json:"n" form:"n"`       //职位名称
	Duly        string `bson:"d" json:"d" form:"d"`       //工作职责
	Requirement string `bson:"r" json:"r" form:"r"`       //岗位要求
	Salary      uint16 `bson:"s" json:"s" form:"s"`       //月薪资
	UploadSize  int64  `bson:"u" json:"u" form:"u"`       //每日上传限额
}

// 查询全部职位
func (this *Job) FindAll() []*Job {
	//查询分类
	var result []*Job
	err := Db.C("job").Find(nil).All(&result)
	if err != nil {
		return nil
	}
	return result
}

// 根据Id查询某个职位信息
func (this *Job) FindOne(id uint8) {
	//查询分类
	err := Db.C("job").Find(bson.M{"_id": id}).One(&this)
	if err != nil {
		return
	}
}

////////////////////////////////////////////////////////////////////////////////
// restful api
////////////////////////////////////////////////////////////////////////////////

// 解析数组
func (this *Job) BaseFormStrings(contro *beego.Controller) (bool, interface{}) {

	return true, nil
}

func (this *Job) BaseCount() int {
	count, err := Db.C("job").Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *Job) BaseFilterCount(filterKey string, filter interface{}) int {
	count, err := Db.C("job").Find(bson.M{filterKey: filter}).Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *Job) BaseInsert() error {
	err := Db.C("job").Insert(this)
	return err
}

func (this *Job) BaseFind(id string) (bool, interface{}) {
	intId, _ := strconv.Atoi(id)
	this.Id = uint8(intId)

	err := Db.C("job").FindId(this.Id).One(&this)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (this *Job) BaseUpdate() error {
	err := Db.C("job").UpdateId(this.Id, this)
	return err
}

func (this *Job) BaseDelete() error {
	err := Db.C("job").RemoveId(this.Id)
	return err
}

/* 设置objectid */
func (this *Job) BaseSetId(id string) (bool, interface{}) {
	intId, _ := strconv.Atoi(id)
	this.Id = uint8(intId)

	return true, nil
}

func (this *Job) BaseSelect(from int, limit int, sort string) []BaseInterface {
	var r []*Job
	err := Db.C("job").Find(nil).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Job) BaseFilterSelect(from int, limit int, sort string, filterKey string, filter interface{}) []BaseInterface {
	var r []*Job
	err := Db.C("job").Find(bson.M{filterKey: filter}).Sort(sort).Skip(from).Limit(limit).All(&r)
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
func (this *Job) BaseSelectLike(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*Job
	err := Db.C("job").Find(bson.M{target: bson.M{"$regex": key}}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Job) BaseLikeCount(target string, key interface{}) int {
	count, err := Db.C("job").Find(bson.M{target: bson.M{"$regex": key}}).Count()
	if err != nil {
		return 0
	}
	return count
}

/* 精确匹配 */
func (this *Job) BaseSelectAccuracy(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*Job

	err := Db.C("job").Find(bson.M{target: key}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Job) BaseAccuracyCount(target string, key interface{}) int {
	count, err := Db.C("job").Find(bson.M{target: key}).Count()
	if err != nil {
		return 0
	}
	return count
}
