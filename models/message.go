package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Message struct {
	Id       bson.ObjectId `bson:"_id"` //主键
	Member   bson.ObjectId `bson:"m"`   //发送给的用户
	Type     string        `bson:"t"`   //消息类型 system(系统消息)
	Info     string        `bson:"i"`   //消息描述
	Category string        `bson:"c"`   //分类
	Read     bool          `bson:"r"`   //是否已阅的
}

var (
	messageC *mgo.Collection //数据库连接
)

func init() {
	messageC = Db.C("message")
}

/* 插入消息 */
func (this *Message) Insert() {
	this.Id = bson.NewObjectId()
	this.Read = false
	err := messageC.Insert(this)
	if err != nil {
		panic(err)
	}
}

/* 查询某个用户的消息
 * @params id 所属讨论id
 * @params from 起始位置
 * @params number 查询数量
 */
func (this *Message) Find(id bson.ObjectId, from int, number int) []*Message {
	var result []*Message
	err := messageC.Find(bson.M{"m": id}).Sort("-_id").Skip(from).Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 查询总页数 */
func (this *Message) Count(id bson.ObjectId) int {
	count, _ := messageC.Find(bson.M{"m": id}).Count()
	return count
}

/* 设置某个消息为已读 */
func (this *Message) SetReaded(memberId bson.ObjectId, id string) bool {
	err := messageC.Update(bson.M{"_id": bson.ObjectIdHex(id), "m": memberId, "r": false}, bson.M{"$set": bson.M{"r": true}})
	if err == nil {
		return true
	} else {
		return false
	}
}

/* 删除某用户200条以后的消息，返回删除的消息中，未读消息有多少 */
func (this *Message) ClearOverMessage(memberId bson.ObjectId) {
	var result []*Message
	err := messageC.Find(bson.M{"m": memberId}).Sort("-_id").Skip(200).All(&result)
	ids := make([]bson.ObjectId, len(result))
	if err != nil {
		return
	}
	for k, _ := range result {
		ids[k] = result[k].Id
	}
	_, err = messageC.RemoveAll(bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return
	}
}
