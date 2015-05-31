package models

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Game struct {
	Id          string        `bson:"_id"` //主键 也是讨论组英文路径
	Manager     bson.ObjectId `bson:"m"`   //管理员
	Managers    []string      `bson:"ms"`  //版主列表 xxx,xxx,xxx （协助版主只能操作帖子，不能进管理台）
	Name        string        `bson:"n"`   //讨论组名称
	Type        uint8         `bson:"t"`   //分类 1~4 智益休闲 养成RPG 竞技类 棋牌类 0为其他，不作为游戏
	Image       []string      `bson:"im"`  //应用截图 最多6个
	Size        float32       `bson:"s"`   //应用大小
	Version     float32       `bson:"v"`   //版本
	Need        string        `bson:"nd"`  //系统要求
	Description string        `bson:"d"`   //简介
	Download    string        `bson:"dl"`  //下载地址
	GameImage   string        `bson:"gi"`  //游戏图标
	Icon        string        `bson:"i"`   //游戏icon图标地址
	Hot         int           `bson:"h"`   //活跃度
	Categorys   int           `bson:"c"`   //分类数
	UpdateTime  time.Time     `bson:"ut"`  //游戏介绍最后更新时间
	Time        time.Time     `bson:"tm"`  //成立时间
}

/* 插入 */
func (this *Game) Insert() bool { //名称或者路径重复则无法插入
	rep := this.FindRepeatName(this.Name)
	if rep {
		return false
	}
	rep = this.FindRepeat(this.Id)
	if rep {
		return false
	}
	//设置时间
	this.UpdateTime = time.Now()
	this.Time = time.Now()
	err := Db.C("game").Insert(this)
	return err == nil
}

/* 更新 */
func (this *Game) Update() bool {
	err := Db.C("game").Update(bson.M{"_id": this.Id}, this)
	return err == nil
}

/* 根据path/id查询某一分类的信息 */
func (this *Game) FindPath(path string) bool {
	//查询游戏
	AutoCache("game-findpath-"+path, &this, 60*60, func() {
		Db.C("game").Find(bson.M{"_id": path}).One(&this)
	})

	return this.Id != ""
}

/* 查询某一段 */
func (this *Game) Find(_type uint8, from int, number int) []*Game {
	var result []*Game
	AutoCache("game-find-"+fmt.Sprintln(_type, from, number), &result, 60*60, func() {
		Db.C("game").Find(bson.M{"t": _type}).Sort("-_id").Skip(from).Limit(number).All(&result)
	})
	return result
}

/* 查询某分类总数 */
func (this *Game) FindCount(_type uint8) int {
	var count int
	AutoCache("game-find-count-"+fmt.Sprintln(_type), &count, 60*60, func() {
		count, _ = Db.C("game").Find(bson.M{"t": _type}).Count()
	})
	return count
}

/* 查询某一段 （游戏） */
func (this *Game) FindGame(from int, number int) []*Game {
	var result []*Game
	AutoCache("game-findgame-"+fmt.Sprintln(from, number), &result, 60*60, func() {
		Db.C("game").Find(bson.M{"t": bson.M{"$ne": 0}}).Sort("-_id").Skip(from).Limit(number).All(&result)
	})
	return result
}

/* 查询前若干个热门（游戏） */
func (this *Game) FindHot(number int) []*Game {
	var result []*Game
	AutoCache("game-findhot"+strconv.Itoa(number), &result, 60*60, func() {
		Db.C("game").Find(bson.M{"t": bson.M{"$ne": 0}}).Sort("-h").Limit(number).All(&result)
	})
	return result
}

/* 查询名称是否重复 */
func (this *Game) FindRepeatName(name string) bool {
	var result *Game
	err := Db.C("game").Find(bson.M{"n": name}).One(&result)
	if err == nil && result != nil {
		return true
	} else {
		return false
	}
}

/* 查询域名是否重复 */
func (this *Game) FindRepeat(id string) bool {
	var result *Game
	err := Db.C("game").Find(bson.M{"_id": id}).One(&result)
	if err == nil && result != nil {
		return true
	} else {
		return false
	}
}

/* 保存全部信息 */
func (this *Game) SaveAll() {
	this.UpdateTime = time.Now()
	Db.C("game").Update(bson.M{"_id": this.Id}, &this)
}

/* 查询全部活跃度小于10并且成立时间超过1个月的游戏 */
func (this *Game) FindBads() []*Game {
	var result []*Game
	weekAgo := time.Now().Add(-30 * 24 * time.Hour)
	err := Db.C("game").Find(bson.M{"h": bson.M{"$lt": 10}, "tm": bson.M{"$lte": weekAgo}}).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 删除 */
func (this *Game) Delete() {
	Db.C("game").RemoveAll(bson.M{"_id": this.Id})
}
