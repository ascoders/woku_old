package models

import (
	"gopkg.in/mgo.v2/bson"
)

type Doc struct {
	Id       bson.ObjectId `bson:"_id" json:"_id"` // 主键 （根文档和Categoryid相同 没有特殊意义） （子文档对应父级DocChild.Id）
	Category bson.ObjectId `bson:"c" json:"c"`     // 所属分类类型【冗余】
	Parent   bson.ObjectId `bson:"p" json:"p"`     // 父级Id（根文档和Categoryid相同 没有特殊意义）
	Childs   []DocChild    `bson:"cs" json:"cs"`   // 子元素
	Nested   int           `bson:"n" json:"n"`     // 嵌套层级
}

type DocChild struct {
	Id       bson.ObjectId `bson:"_id" json:"_id"` // 主键 文章对应topicid 文件夹没有特殊含义（如果有展开的文件夹行，对应那行的Doc.Id）
	IsFolder bool          `bson:"if" json:"ifr"`  // 是否为文件夹
	Name     string        `bson:"n" json:"n"`     // 文件名
}

// 根据id查询
func (this *Doc) Find(id string) bool {
	if !bson.IsObjectIdHex(id) {
		return false
	}

	err := Db.C("doc").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&this)

	return err == nil
}

// 插入新文档
func (this *Doc) Insert(id string, category string, parent string) bool {
	if !bson.IsObjectIdHex(id) {
		return false
	}

	if !bson.IsObjectIdHex(category) {
		return false
	}

	if !bson.IsObjectIdHex(parent) {
		return false
	}

	this.Id = bson.ObjectIdHex(id)
	this.Category = bson.ObjectIdHex(category)
	this.Parent = bson.ObjectIdHex(parent)

	err := Db.C("doc").Insert(this)

	return err == nil
}

// 对某个文尾部部插入子元素
func (this *Doc) PushChild(child DocChild) bool {
	err := Db.C("doc").Update(bson.M{"_id": this.Id},
		bson.M{"$push": bson.M{
			"cs": child,
		}})

	return err == nil
}

// 删除某个子文档
func (this *Doc) DeleteChild(child DocChild) bool {
	err := Db.C("doc").Update(bson.M{"_id": this.Id},
		bson.M{"$pull": bson.M{
			"cs": child,
		}})

	return err == nil
}

// 查询拥有某个文章id的文档
func (this *Doc) IncludeChild(childId string) bool {
	if !bson.IsObjectIdHex(childId) {
		return false
	}

	err := Db.C("doc").Find(bson.M{"cs._id": bson.ObjectIdHex(childId)}).One(&this)

	return err == nil
}

// 更新子文档
func (this *Doc) UpdateChild(childs []DocChild) bool {
	var err error

	err = Db.C("doc").Update(bson.M{"_id": this.Id},
		bson.M{"$set": bson.M{
			"cs": childs,
		}})

	return err == nil
}
