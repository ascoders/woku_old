package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Order struct {
	Id            bson.ObjectId `bson:"_id"` //主键
	ToId          bson.ObjectId `bson:"ti"`  //充值给的账户ID
	ToName        string        `bson:"tn"`  //充值给的账户昵称
	FromId        bson.ObjectId `bson:"fi"`  //付款的账户ID
	FromName      string        `bson:"fn"`  //付款的账户昵称
	Pay           float32       `bson:"p"`   //付款金额
	Success       bool          `bson:"s"`   //是否成功付款
	PayPlantform  string        `bson:"pf"`  //支付平台 (alipay)
	AlipayAccount string        `bson:"aa"`  //卖家支付宝账号
	AlipayNumber  string        `bson:"an"`  //支付平台唯一订单号(交易成功后)
	Time          time.Time     `bson:"t"`   //订单交易成功日期
	Type          string        `bson:"tp"`  //订单类型 web
	Description   string        `bson:"d"`   //描述
	Gain          string        `bson:"g"`   //获取商品
	Notify        string        `bson:"n"`   //回调地址
	Reqid         string        `bson:"r"`   //自定义充值账户唯一ID
	Game          string        `bson:"gm"`  //游戏路径
	Extend        string        `bson:"e"`   //拓展参数
}

var (
	orderC *mgo.Collection //数据库连接
)

func init() {
	orderC = Db.C("order")
}

/* 创建一个待付款账单 */
func (this *Order) InsertOrder() {
	this.Id = bson.NewObjectId()
	this.Success = false
	this.Time = bson.Now()

	err := orderC.Insert(this)
	if err != nil {
		return
	} else {
		return
	}
}

/* 查询某个订单 */
func (this *Order) FindOne() bool {
	//查询订单
	err := orderC.Find(bson.M{"_id": this.Id}).One(&this)
	if err != nil {
		return false
	}
	return true
}

/* 更新订单信息 */
func (this *Order) Update(change bson.M) bool {
	//加入乐观锁机制，仅当数据库success为false时才允许更新
	colQuerier := bson.M{"_id": this.Id, "s": false}
	err := orderC.Update(colQuerier, change)
	if err != nil { //更新出错
		return false
	} else {
		return true
	}
}

/* 查询某个用户的订单信息 */
func (this *Order) Count(memberId bson.ObjectId) int {
	count, _ := orderC.Find(bson.M{"fi": memberId}).Count()
	return count
}

/* 查询某个用户的订单
 * @params id 所属讨论id
 * @params from 起始位置
 * @params number 查询数量
 */
func (this *Order) Find(id bson.ObjectId, from int, number int) []*Order {
	var result []*Order
	err := orderC.Find(bson.M{"fi": id}).Sort("-_id").Skip(from).Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 删除时间超过2小时但未支付的订单 */
func (this *Order) DeleteBad() {
	twoHourAgo := time.Now().Add(-2 * time.Hour)
	orderC.RemoveAll(bson.M{"t": bson.M{"$lte": twoHourAgo}, "s": false})
}
