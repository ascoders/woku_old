package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 常用索引
// Id
// Game + Category

type Topic struct {
	Id            bson.ObjectId `bson:"_id"` // 主键
	Author        bson.ObjectId `bson:"a"`   // 作者id
	AuthorName    string        `bson:"an"`  // 作者昵称
	AuthorImage   string        `bson:"ai"`  // 作者头像地址
	LastReply     bson.ObjectId `bson:"l"`   // 最后一个回帖者id
	LastReplyName string        `bson:"ln"`  // 最后一个回帖者昵称
	LastTime      time.Time     `bson:"lt"`  // 最后一次回帖时间
	Game          string        `bson:"g"`   // 所属game
	Ip            string        `bson:"i"`   // 发帖者IP地址
	Category      bson.ObjectId `bson:"cg"`  // 所属分类
	Title         string        `bson:"t"`   // 话题标题，可空
	Content       string        `bson:"c"`   // 内容
	ContentSego   []string      `bson:"cs"`  // 内容分词表
	Time          time.Time     `bson:"tm"`  // 发表日期
	Reply         int           `bson:"r"`   // 回复总数（包含嵌套评论）
	OutReply      int           `bson:"o"`   // 回复总数（只包含最外层回复）
	Star          int           `bson:"s"`   // 评分
	Views         int           `bson:"v"`   // 浏览数
	Good          bool          `bson:"gd"`  // 加精 true表示加精
	Tag           []string      `bson:"ta"`  // 标签
	Top           int           `bson:"tp"`  // 置顶 0表示不置顶
}

var (
	topicC *mgo.Collection //数据库连接
)

func init() {
	topicC = Db.C("topic")
}

/* 设置一个id */
func (this *Topic) SetId() {
	this.Id = bson.NewObjectId()
}

/* 插入话题 */
func (this *Topic) Insert() bson.ObjectId {

	this.Time = time.Now()
	this.LastTime = this.Time

	err := topicC.Insert(this)
	if err != nil {
		panic(err)
	}
	return this.Id
}

/* 查询指定游戏一定范围的话题
 * @params game 所属游戏id
 * @params category 所属分类
 * @params from 起始位置
 * @params number 查询数量
 */
func (this *Topic) Find(game string, category string, from int, number int) []*Topic {
	var tops []*Topic
	var result []*Topic

	if !bson.IsObjectIdHex(category) {
		return nil
	}

	//如果查询第一页，查询置顶
	if from == 0 {
		topicC.Find(bson.M{"g": game, "cg": bson.ObjectIdHex(category), "tp": bson.M{"$gt": 0}}).Sort("-tp").All(&tops)
	}
	//查询指定范围内帖子
	err := topicC.Find(bson.M{"g": game, "cg": bson.ObjectIdHex(category), "tp": 0}).Sort("-lt").Skip(from).Limit(number).All(&result)
	//置顶帖后追加普通帖子
	tops = append(tops, result...)
	if err != nil {
		return nil
	}
	return tops
}

func (this *Topic) FindByTag(game string, tag string, from int, number int) []*Topic {
	var result []*Topic

	//查询指定范围内帖子
	err := topicC.Find(bson.M{"g": game, "ta": tag}).Sort("-lt").Skip(from).Limit(number).All(&result)

	if err != nil {
		return nil
	}
	return result
}

/* 查询最新的N个帖子 */
func (this *Topic) FindNew(game string, category bson.ObjectId, number int) []*Topic {
	var result []*Topic
	//查询指定范围内帖子
	err := topicC.Find(bson.M{"g": game, "cg": category, "tp": 0}).Sort("-lt").Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 查询指定游戏某分类的话题数 */
func (this *Topic) FindCount(game string, category string) int {
	if !bson.IsObjectIdHex(category) {
		return 0
	}

	count, err := topicC.Find(bson.M{"g": game, "cg": bson.ObjectIdHex(category), "tp": 0}).Count()
	if err == nil {
		return int(count)
	} else {
		return 0
	}
}

/* 查询指定游戏某标签的话题数 */
func (this *Topic) FindTagCount(game string, tag string) int {
	count, err := topicC.Find(bson.M{"g": game, "ta": tag}).Count()
	if err == nil {
		return int(count)
	} else {
		return 0
	}
}

/* 查询指定游戏话题具体内容 */
func (this *Topic) FindOne(game string, id string) bool {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(id) {
		return false
	}
	//查询话题
	err := topicC.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&this)
	if err == nil && this.Game == game { //查询到了并且是此game下的话题
		return true
	} else {
		return false
	}
}

/* 查询某话题信息 */
func (this *Topic) FindById(id string) bool {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(id) {
		return false
	}
	//查询话题
	err := topicC.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&this)

	if err == nil {
		return true
	} else {
		return false
	}
}

/* 话题浏览数量加1 */
func (this *Topic) AddViews(id string) {
	//如果不是bsonid则退出
	if !bson.IsObjectIdHex(id) {
		return
	}

	topicC.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$inc": bson.M{
			"v": 1,
		}})
}

/* 给某个id文章置顶 */
func (this *Topic) AddTop(id string, game string, category bson.ObjectId, maxNumber int) (bool, int) {
	//先查询置顶文章数量
	number, _ := topicC.Find(bson.M{"g": game, "cg": category, "tp": bson.M{"$gt": 0}}).Count()
	if number >= maxNumber { //超过最大限度了
		return false, 0
	}
	//查询置顶文章中tp最大的那个
	var maxTp *Topic
	max := 0
	err := topicC.Find(bson.M{"g": game, "cg": category, "tp": bson.M{"$gt": 0}}).Sort("-tp").One(&maxTp)
	if err == nil && maxTp != nil {
		max = maxTp.Top
	}
	//将该id文章tp设置为最大+1
	err = topicC.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"tp": max + 1,
		}})
	if err != nil {
		return false, 0
	} else {
		return true, max + 1
	}
}

/* 文档 文件数加1 */
func (this *Topic) AddCountNumber(id bson.ObjectId) {
	topicC.Update(bson.M{"_id": id},
		bson.M{"$inc": bson.M{
			"cn": 1,
		}})
}

/* 取消某id的置顶 */
func (this *Topic) CancleTop(id string) bool {
	//将该id文章tp设置为0
	err := topicC.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"tp": 0,
		}})
	if err != nil {
		return false
	} else {
		return true
	}
}

/* 给某个id文章加精 */
func (this *Topic) AddGood(id string) bool {
	err := topicC.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"gd": true,
		}})
	if err != nil {
		return false
	} else {
		return true
	}
}

/* 给某个id文章取消精华 */
func (this *Topic) CancleGood(id string) bool {
	err := topicC.Update(bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"$set": bson.M{
			"gd": false,
		}})
	if err != nil {
		return false
	} else {
		return true
	}
}

/* 删除某个游戏下全部话题 */
func (this *Topic) DeleteGameTopic(game string) {
	_, err := topicC.RemoveAll(bson.M{"g": game})
	if err != nil {
		return
	}
}

/* 删除某个id的文章 */
func (this *Topic) Delete() bool {
	err := topicC.Remove(bson.M{"_id": this.Id})
	if err != nil {
		return false
	} else {
		return true
	}
}

/* 保存某话题 */
func (this *Topic) Save() {
	err := topicC.Update(bson.M{"_id": this.Id}, this)
	if err != nil {
		return
	}
}

/* 查询本周前10热门论坛帖子 */
func (this *Topic) WeekTopBbs() []*Topic {
	var result []*Topic
	if cache, err := Redis.Get("weektop"); err == nil {
		StructDecode(cache, &this)
	} else {
		weekAgo := time.Now().Add(-7 * 24 * time.Hour)
		topicC.Find(bson.M{"cg": "bbs", "tm": bson.M{"$gte": weekAgo}}).Sort("-r").Limit(10).All(&result)
		if length := len(result); length < 10 {
			need := 10 - length
			var extra []*Topic
			err := topicC.Find(bson.M{"cg": "bbs", "tm": bson.M{"$lt": weekAgo}}).Sort("-r").Limit(need).All(&extra)
			if err == nil {
				result = append(result, extra...)
			}
		}
		cache, _ := StructEncode(this)
		Redis.Setex("weektop-", 60*60*24, cache)
	}
	return result
}

// 查询某游戏某分类下是否有文章
func (this *Topic) HasTopic(game string, category string) bool {
	if !bson.IsObjectIdHex(category) {
		return false
	}

	err := topicC.Find(bson.M{"g": game, "cg": bson.ObjectIdHex(category)}).One(nil)
	return err == nil
}

// 新增标签
func (this *Topic) AddTag(name string) {
	Db.C("topic").Update(bson.M{"_id": this.Id},
		bson.M{"$push": bson.M{
			"ta": name,
		}})
}

// 删除标签
func (this *Topic) RemoveTag(name string) {
	Db.C("topic").Update(bson.M{"_id": this.Id},
		bson.M{"$pull": bson.M{
			"ta": name,
		}})
}

// 查询拥有相同标签的文章
func (this *Topic) Same() []*Topic {
	var r []*Topic

	Db.C("topic").Find(bson.M{"ta": bson.M{"$in": this.Tag}, "_id": bson.M{"$ne": this.Id}}).Limit(10).All(&r)

	return r
}
