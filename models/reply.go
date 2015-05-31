package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Reply struct {
	Id          bson.ObjectId `bson:"_id"` //主键
	Game        string        `bson:"g"`   //所属游戏
	Topic       string        `bson:"t"`   //所属帖子
	Reply       string        `bson:"rp"`  //所属父级回复（仅当自己为嵌套回复）
	Author      bson.ObjectId `bson:"a"`   //作者id
	AuthorName  string        `bson:"an"`  //作者昵称
	AuthorImage string        `bson:"ai"`  //作者头像地址
	Ip          string        `bson:"i"`   //作者IP地址
	Content     string        `bson:"c"`   //内容
	Time        time.Time     `bson:"tm"`  //发表日期
	ReplyNumber int           `bson:"r"`   //回复数量
	ReplyCache  []Reply       `bson:"rt"`  //回复缓存 缓存3个回复
}

var (
	replyC *mgo.Collection //数据库连接
)

func init() {
	replyC = Db.C("reply")
}

/* 设置id */
func (this *Reply) SetId() {
	this.Id = bson.NewObjectId()
}

/* 插入回复 */
func (this *Reply) Insert() {
	this.Time = time.Now()

	err := replyC.Insert(this)
	if err != nil {
		return
	}
}

/* 查询某个话题下一定数量的回复
 * @params topicId 所属讨论id
 * @params from 起始位置
 * @params number 查询数量
 */
func (this *Reply) Find(topicId string, from int, number int) []*Reply {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(topicId) {
		return nil
	}

	var result []*Reply
	err := replyC.Find(bson.M{"t": topicId, "rp": ""}).Sort("_id").Skip(from).Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 查询回复 */
func (this *Reply) FindById(id string) bool {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(id) {
		return false
	}
	//查询话题
	err := replyC.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&this)
	if err == nil {
		return true
	} else {
		return false
	}
}

/* 保存 */
func (this *Reply) Save() {
	err := replyC.Update(bson.M{"_id": this.Id}, &this)
	if err != nil {
		return
	}
}

/* 删除某个游戏下全部回复和评论 */
func (this *Reply) DeleteGameReply(game string) {
	_, err := replyC.RemoveAll(bson.M{"g": game})
	if err != nil {
		return
	}
}

/* 删除某个帖子下全部回复和评论 */
func (this *Reply) DeleteTopicReply(topic string) {
	_, err := replyC.RemoveAll(bson.M{"t": topic})
	if err != nil {
		return
	}
}

/* 删除某个回复下的全部评论 */
func (this *Reply) DeleteReplyReply(reply string) *mgo.ChangeInfo {
	change, err := replyC.RemoveAll(bson.M{"rp": reply})
	if err != nil {
		return nil
	}
	return change
}

/* 删除 */
func (this *Reply) Delete() {
	_, err := replyC.RemoveAll(bson.M{"_id": this.Id})
	if err != nil {
		return
	}
}

/* 刷新前五个嵌套评论 */
func (this *Reply) FreshCache() {
	var result []Reply
	replyC.Find(bson.M{"rp": this.Id.Hex(), "_id": bson.M{"$ne": this.Id}}).Limit(5).Sort("_id").All(&result)
	this.ReplyCache = result
}

/* 查询某个回复下一定数量的回复
 * @params replyId 所属回复id
 * @params from 起始位置
 * @params number 查询数量
 */
func (this *Reply) FindReply(replyId string, from int, number int) []*Reply {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(replyId) {
		return nil
	}

	var result []*Reply
	err := replyC.Find(bson.M{"rp": replyId}).Sort("_id").Skip(from).Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}
