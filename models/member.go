package models

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Member struct {
	Id            bson.ObjectId `bson:"_id" json:"_id"`         //主键
	Email         string        `bson:"e" json:"e" form:"e"`    //电子邮箱 也是注册时候的账号
	Mobile        string        `bson:"m" json:"m" form:"m"`    //手机号
	Password      string        `bson:"p" json:"p" form:"p"`    //密码
	Nickname      string        `bson:"n" json:"n" form:"n"`    //昵称
	Sex           uint8         `bson:"s" json:"s" form:"s"`    //性别
	Money         float32       `bson:"mo" json:"mo" form:"mo"` //账户余额
	MoneyHistory  string        `bson:"mh" json:"mh" form:"mh"` //账户余额历史（按天计算）
	MonthFree     int           `bson:"mf" json:"mf" form:"mf"` //每月免费额度
	LogTime       uint16        `bson:"l" json:"l" form:"l"`    //登陆次数
	LastTime      time.Time     `bson:"la" json:"la" form:"la"` //最后操作时间
	ErrorChance   uint8         `bson:"er" json:"er" form:"er"` //账号输错机会次数
	StopTime      time.Time     `bson:"st" json:"st" form:"st"` //账号封停截至时间
	Type          uint8         `bson:"t" json:"t" form:"t"`    //账号类型 0:超级管理员/董事长 1:会员 2:高级会员 3:白金会员
	Power         []string      `bson:"po" json:"po" form:"po"` //权限范围
	SalayTime     time.Time     `bson:"sa" json:"sa" form:"sa"` //最后领取工资的时间
	UploadSize    int64         `bson:"u" json:"u" form:"u"`    //当前上传大小
	UploadDate    time.Time     `bson:"ud" json:"ud" form:"ud"` //最后上传文件的时间
	LockVersion   uint64        `bson:"lv" json:"lv" form:"lv"` //乐观锁
	HasOrder      bool          `bson:"h" json:"h" form:"h"`    //是否有未处理的账单
	Token         string        `bson:"tk" json:"tk" form:"tk"` //每个账号的密钥
	Image         string        `bson:"i" json:"i" form:"i"`    //头像地址
	MessageNumber int           `bson:"mn" json:"mn" form:"mn"` //未读消息数量
	MessageAll    int           `bson:"ma" json:"ma" form:"ma"` //总消息数
	GameNumber    int           `bson:"g" json:"g" form:"g"`    //建立游戏数量
}

var (
	memberC *mgo.Collection //数据库连接
)

func init() {
	memberC = Db.C("member")
}

/* 插入用户 */
func (this *Member) Insert() string {
	if this.Id == "" { //id没有设置则初始化新的
		this.Id = bson.NewObjectId()
	}
	this.LastTime = bson.Now()
	this.StopTime = bson.Now()
	this.SalayTime = bson.Now()
	this.Token = strconv.Itoa(int(rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()))
	this.ErrorChance = 6
	//如果没有头像，随机设置一个头像
	if this.Image == "" {
		this.Image = strconv.Itoa(rand.Intn(9))
	}
	//密码md5加盐+用户ID+md5加密
	m := md5.New()
	m.Write([]byte(this.Password))
	n := md5.New()
	n.Write([]byte(hex.EncodeToString(m.Sum(nil)) + beego.AppConfig.String("md5salt") + bson.ObjectId.Hex(this.Id)))
	this.Password = hex.EncodeToString(n.Sum(nil))
	err := memberC.Insert(this)
	if err != nil {
		panic(err)
	}
	return bson.ObjectId.Hex(this.Id)
}

/* 更新用户信息 */
func (this *Member) Update(change bson.M) bool {
	//加入乐观锁机制，仅当change更新字段有更新乐观锁时才会启用
	//若change中不含有乐观锁字段+1更新，每次查询必定有结果
	//如果某个更新可能产生并发问题，一定要更新乐观锁，这样下次更新
	//更新可能没有结果，返回err，重新更新
	colQuerier := bson.M{"_id": this.Id, "lv": this.LockVersion}
	err := memberC.Update(colQuerier, change)
	if err != nil { //更新出错
		return false
	} else {
		return true
	}
}

/* 更新其他用户信息 */
func (this *Member) UpdateOther(id bson.ObjectId, change bson.M) bool {
	colQuerier := bson.M{"_id": id}
	err := memberC.Update(colQuerier, change)
	if err != nil {
		return false
	}
	return true
}

/* 根据ObjectId查询某个用户信息 */
func (this *Member) FindOne(value string) bool {
	//检查id格式是否正确
	if !bson.IsObjectIdHex(value) {
		return false
	}
	//查询用户
	err := memberC.Find(bson.M{"_id": bson.ObjectIdHex(value)}).One(&this)
	if err != nil {
		return false
	}
	return true
}

/* 根据账号查询某个用户信息 支持类型：邮箱 */
func (this *Member) FindByAccount(account string) bool {
	//查询用户
	err := memberC.Find(bson.M{"e": account}).One(&this)
	if err != nil {
		return false
	}
	return true
}

/* 根据QQ openid查找用户 */
func (this *Member) FindByQQ(openid string) bool {
	//查询用户
	err := memberC.Find(bson.M{"oq": openid}).One(&this)
	if err != nil {
		return false
	}
	return true
}

/* 根据昵称查找用户 */
func (this *Member) FindByNickname(nickname string) bool {
	//查询用户
	err := memberC.Find(bson.M{"n": nickname}).One(&this)
	if err != nil {
		return false
	}
	return true
}

/* 修改密码 */
func (this *Member) ChangePassword(password string) {
	//密码加密
	m := md5.New()
	m.Write([]byte(password))
	n := md5.New()
	n.Write([]byte(hex.EncodeToString(m.Sum(nil)) + beego.AppConfig.String("md5salt") + bson.ObjectId.Hex(this.Id)))

	//密码赋值
	this.Password = hex.EncodeToString(n.Sum(nil))
	this.Update(bson.M{"$set": bson.M{"p": this.Password}})
}

/* 校验密码是否正确 支持类型：邮箱 */
func (this *Member) CheckPass(account string, password string) (string, int) {
	err := memberC.Find(bson.M{"e": account}).One(&this)
	if err != nil {
		return "", -2 //账号不存在
	}

	if bson.Now().Before(this.StopTime) { //锁定时间还没过
		long := this.StopTime.Sub(bson.Now())
		return strconv.FormatFloat(long.Seconds(), 'f', 0, 64) + " 秒", -3
	}
	//密码加密
	m := md5.New()
	m.Write([]byte(password))
	n := md5.New()
	n.Write([]byte(hex.EncodeToString(m.Sum(nil)) + beego.AppConfig.String("md5salt") + bson.ObjectId.Hex(this.Id)))
	//对比
	if hex.EncodeToString(n.Sum(nil)) != this.Password { //验证出错
		if this.ErrorChance <= 1 { //用尽验证机会，账号锁定10分钟
			this.ErrorChance = 6
			minute := time.Duration(10) * time.Minute
			this.StopTime = bson.Now().Add(minute)
			this.Update(bson.M{"$set": bson.M{"er": this.ErrorChance, "st": this.StopTime}})
			return "", -4 //进入锁定
		} else { //验证机会-1
			this.ErrorChance--
			this.Update(bson.M{"$set": bson.M{"er": this.ErrorChance}})
			return strconv.Itoa(int(this.ErrorChance)), -1 //密码不匹配
		}
	} else { //通过验证，重置机会次数
		this.ErrorChance = 6
		this.Update(bson.M{"$set": bson.M{"er": this.ErrorChance}})
		return this.Id.Hex(), 1
	}
}

/* 查询邮箱是否已存在 */
func (this *Member) EmailExist(email string) bson.ObjectId {
	var member *Member
	err := memberC.Find(bson.M{"e": email}).One(&member)
	if err != nil {
		return ""
	}
	return member.Id
}

/* 查询昵称是否已存在 */
func (this *Member) NicknameExist(nickname string) bson.ObjectId {
	var member *Member
	err := memberC.Find(bson.M{"n": nickname}).One(&member)
	if err != nil {
		return ""
	}
	return member.Id
}

/* 查询qq的openid是否已存在 */
func (this *Member) QQOpenIdExist(openid string) bson.ObjectId {
	var member *Member
	err := memberC.Find(bson.M{"oq": openid}).One(&member)
	if err != nil {
		return ""
	}
	return member.Id
}

/* 更新用户信息 */
func (this *Member) UpdateFinish() {
	colQuerier := bson.M{"_id": this.Id}
	change := bson.M{"$set": bson.M{"la": bson.Now(), "l": this.LogTime}}
	err := memberC.Update(colQuerier, change)
	if err != nil {
		//处理错误
	}
}

/* 查询某个职位下是否有用户 */
func (this *Member) TypeHasMember(_type uint8) bool {
	var result []*Member
	err := memberC.Find(bson.M{"t": _type}).All(&result)
	if err != nil {
		//处理错误
		return false
	}
	if result == nil {
		return false
	} else {
		return true
	}
}

/* 查询账号总数
 * @params special bool 是否只查公司特殊权限账号
 */
func (this *Member) Count(special bool) int {
	findQuery := bson.M{}
	if special {
		findQuery = bson.M{"t": bson.M{"$gt": 3}}
	}
	count, err := memberC.Find(findQuery).Count()
	if err != nil {
		return 0
	}
	return count
}

/* 查询部分账号
 * @params special bool 是否只查公司特殊权限账号
 * @params from int 开始位置
 * @params number int 查询数量
 */
func (this *Member) Find(special bool, from int, number int) []*Member {
	var result []*Member
	findQuery := bson.M{}
	if special {
		findQuery = bson.M{"t": bson.M{"$gt": 3}}
	}
	err := memberC.Find(findQuery).Sort("_id").Skip(from).Limit(number).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 生成签名 md5(id&expire&token&type)
 * @params expire int64 过期时间
 * @params _type string 操作类型
 * @params params string 其他参数
 * @return 签名结果 url参数
 */
func (this *Member) CreateSign(expire int64, _type string, extend string) (string, string) {
	endTime := strconv.Itoa(int(expire))
	sign := this.Id.Hex() + endTime + _type + extend + this.Token
	//sign md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign, "?id=" + this.Id.Hex() + "&expire=" + endTime + "&type=" + _type + "&extend=" + extend + "&sign=" + sign
}

/* 模糊查找 */
func (this *Member) FindLike(key string) []*Member {
	var result []*Member
	err := memberC.Find(bson.M{"n": bson.M{"$regex": key}}).Limit(5).All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 保存 */
func (this *Member) Save() {
	colQuerier := bson.M{"_id": this.Id}
	err := memberC.Update(colQuerier, this)
	if err != nil {
		//处理错误
	}
}

////////////////////////////////////////////////////////////////////////////////
// restful api
////////////////////////////////////////////////////////////////////////////////

// 解析数组
func (this *Member) BaseFormStrings(contro *beego.Controller) (bool, interface{}) {
	this.Power = contro.GetStrings("po")

	return true, nil
}

func (this *Member) BaseCount() int {
	count, err := Db.C("member").Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *Member) BaseFilterCount(filterKey string, filter interface{}) int {
	count, err := Db.C("member").Find(bson.M{filterKey: filter}).Count()
	if err != nil {
		return 0
	}
	return count
}

func (this *Member) BaseInsert() error {
	return nil
}

func (this *Member) BaseFind(id string) (bool, interface{}) {
	if !bson.IsObjectIdHex(id) {
		return false, "id类型错误"
	}

	err := Db.C("member").FindId(bson.ObjectIdHex(id)).One(&this)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (this *Member) BaseUpdate() error {
	err := Db.C("member").UpdateId(this.Id, this)
	return err
}

func (this *Member) BaseDelete() error {
	return nil
}

/* 设置objectid */
func (this *Member) BaseSetId(id string) (bool, interface{}) {
	if !bson.IsObjectIdHex(id) {
		return false, "id类型错误"
	}

	this.Id = bson.ObjectIdHex(id)
	return true, nil
}

func (this *Member) BaseSelect(from int, limit int, sort string) []BaseInterface {
	var r []*Member
	err := Db.C("member").Find(nil).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Member) BaseFilterSelect(from int, limit int, sort string, filterKey string, filter interface{}) []BaseInterface {
	var r []*Member
	err := Db.C("member").Find(bson.M{filterKey: filter}).Sort(sort).Skip(from).Limit(limit).All(&r)
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
func (this *Member) BaseSelectLike(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*Member
	err := Db.C("member").Find(bson.M{target: bson.M{"$regex": key}}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Member) BaseLikeCount(target string, key interface{}) int {
	count, err := Db.C("member").Find(bson.M{target: bson.M{"$regex": key}}).Count()
	if err != nil {
		return 0
	}
	return count
}

/* 精确匹配 */
func (this *Member) BaseSelectAccuracy(from int, limit int, sort string, target string, key interface{}) []BaseInterface {
	var r []*Member

	err := Db.C("member").Find(bson.M{target: key}).Sort(sort).Skip(from).Limit(limit).All(&r)
	if err != nil {
		return nil
	}
	result := make([]BaseInterface, len(r))
	for k, _ := range result {
		result[k] = r[k]
	}
	return result
}

func (this *Member) BaseAccuracyCount(target string, key interface{}) int {
	count, err := Db.C("member").Find(bson.M{target: key}).Count()
	if err != nil {
		return 0
	}
	return count
}
