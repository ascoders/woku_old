package models

import (
	"gopkg.in/mgo.v2/bson"
)

type GameCategory struct {
	Id           bson.ObjectId `bson:"_id"` // 主键
	Game         string        `bson:"g"`   // 所属游戏
	Category     string        `bson:"c"`   // 分类名 也是讨论组英文路径
	CategoryName string        `bson:"cn"`  // 分类中文名
	Icon         string        `bson:"i"`   // 图标 fontAwesome
	Recommend    int           `bson:"re"`  // 推荐数量
	RecommendPri int           `bson:"rp"`  // 推荐优先级
	Add          int           `bson:"a"`   // 发帖策略 0:只有管理员和协助管理员可以回帖 1:登陆用户可回帖
	Reply        int           `bson:"r"`   // 回帖策略 同发帖
	Number       int           `bson:"n"`   // 文章数
	Type         int           `bson:"t"`   // 分类 0:论坛 1:文档
}

/* 插入 */
func (this *GameCategory) Insert() bool {
	//同一个游戏不能有重复分类和分类名称
	err := Db.C("gameCategory").Find(bson.M{"g": this.Game, "c": this.Category}).One(nil)
	if err == nil { //有重复的则退出
		return false
	}

	this.Id = bson.NewObjectId()
	err = Db.C("gameCategory").Insert(this)
	return err == nil
}

/* 查询某个游戏的全部分类 */
func (this *GameCategory) FindCategorys(game string) []*GameCategory {
	var result []*GameCategory
	Db.C("gameCategory").Find(bson.M{"g": game}).All(&result)
	return result
}

/* 查询游戏的某个分类 */
func (this *GameCategory) Find(game string, category string) bool {
	if !bson.IsObjectIdHex(category) {
		return false
	}

	err := Db.C("gameCategory").Find(bson.M{"g": game, "_id": bson.ObjectIdHex(category)}).One(&this)
	return err == nil
}

// 查询某个分类
func (this *GameCategory) FindId(category string) bool {
	if !bson.IsObjectIdHex(category) {
		return false
	}

	err := Db.C("gameCategory").Find(bson.M{"_id": bson.ObjectIdHex(category)}).One(&this)
	return err == nil
}

// 更新某个游戏某个分类优先级
func (this *GameCategory) ChangeRecommendPri(game string, category string, value int) (bool, interface{}) {
	err := Db.C("gameCategory").Update(bson.M{"g": game, "c": category}, bson.M{"$set": bson.M{"rp": value}})
	return err == nil, err
}

// 更新某游戏某分类各项数值
func (this *GameCategory) Update(id string, path string, name string, number int, add int, reply int, _type int) (bool, interface{}) {
	if !bson.IsObjectIdHex(id) {
		return false, "id格式错误"
	}

	err := Db.C("gameCategory").Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{"c": path, "cn": name, "re": number, "a": add, "r": reply, "t": _type}})
	return err == nil, err
}

// 删除某个分类
func (this *GameCategory) Delete(game string, id string) (bool, interface{}) {
	if !bson.IsObjectIdHex(id) {
		return false, "id格式错误"
	}

	err := Db.C("gameCategory").Remove(bson.M{"g": game, "_id": bson.ObjectIdHex(id)})
	return err == nil, err
}
