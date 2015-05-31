package models

import (
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Web struct {
	Id           uint8  `bson:"_id"` //主键
	ViewsHistory string `bson:"vh"`  //总浏览量历史
}

var (
	webC *mgo.Collection //数据库连接
)

func init() {
	webC = Db.C("web")
}

/* 获取网站信息 */
func (this *Web) GetInfo() {
	err := webC.Find(bson.M{"_id": 0}).One(&this)
	if err != nil {
		return
	}
}

/* 追加总浏览量历史 */
func (this *Web) AddViewHistory(views uint64) {
	if this.ViewsHistory == "" {
		this.ViewsHistory = strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.Itoa(int(views))
	} else {
		this.ViewsHistory += ";" + strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.Itoa(int(views))
	}
	historyArray := strings.Split(this.ViewsHistory, ";")
	if len(historyArray) > 30 { //如果历史记录超过30条
		//只取前30条
		subArray := historyArray[len(historyArray)-30 : len(historyArray)]
		this.ViewsHistory = strings.Join(subArray, ";")
	}
	err := webC.Update(bson.M{"_id": 0}, bson.M{"$set": bson.M{"vh": this.ViewsHistory}})
	if err != nil {
		return
	}
}
