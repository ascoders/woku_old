package controllers

import (
	"regexp"
	"strconv"
	"time"
	"woku/models"

	"github.com/ascoders/alipay"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	beego.Controller
	member models.Member //用户
}

var (
	UserLog *logs.BeeLogger //打印日志
)

func init() {
	//初始化日志
	UserLog = logs.NewLogger(10000)
	UserLog.SetLogger("file", `{"filename":"log/user.log"}`)
}

func (this *UserController) Prepare() {
	//获取用户信息
	ok := false
	if session := this.GetSession("WOKUID"); session != nil {
		if _ok := this.member.FindOne(session.(string)); _ok {
			ok = true
		}
	}

	if !ok { //用户不存在，退出
		this.Data["json"] = map[string]interface{}{
			"ok":   false,
			"data": "未登录",
		}
		this.ServeJson()
		this.StopRun()
	}
}

/* 获取消息列表 */
func (this *UserController) GetMessage() {
	from, _ := this.GetInt("from")
	number, _ := this.GetInt("number")

	ok, data := func() (bool, interface{}) {
		if number > 100 {
			return false, "最多显示100项"
		}

		message := &models.Message{}
		messages := message.Find(this.member.Id, from, number)

		//获取消息总数
		count := message.Count(this.member.Id)

		//清空用户未读消息数
		this.member.MessageNumber = 0
		this.member.Save()

		return true, map[string]interface{}{
			"lists": messages,
			"count": count,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 消息已读请求-post */
func (this *UserController) AccountMessageReadPost() {
	message := &models.Message{}
	message.SetReaded(this.member.Id, this.GetString("id"))
}

/* 获取当前未读消息数量-post */
func (this *UserController) AccountMessageNumber() {
	this.Data["json"] = this.member.MessageNumber
	this.ServeJson()
}

/* 充值记录 */
func (this *UserController) GetHistory() {
	from, _ := this.GetInt("from")
	number, _ := this.GetInt("number")

	ok, data := func() (bool, interface{}) {
		if number > 100 {
			return false, "最多显示100项"
		}

		order := &models.Order{}
		orders := order.Find(this.member.Id, from, number)

		// 查询总页数
		count := order.Count(this.member.Id)

		return true, map[string]interface{}{
			"lists": orders,
			"count": count,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 修改密码 */
func (this *UserController) Password() {
	ok, data := func() (bool, interface{}) {
		if ok, err := regexp.MatchString("[0-9a-zA-Z_-]{6,30}", this.GetString("password")); !ok || err != nil {
			return false, "密码长度为6~30的字母或数字"
		}

		this.member.ChangePassword(this.GetString("password"))
		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 账户充值-post 处理充值请求 */
func (this *UserController) Recharge() {
	ok, data := func() (bool, interface{}) {
		if _, err := models.Redis.Get(this.member.Id.Hex() + "recharge"); err == nil {
			return false, "您操作过于频繁，10秒后再试"
		}

		//验证数据完整性
		if this.GetString("account") == "" {
			return false, "充值账户必须填写"
		}
		number, _ := this.GetFloat("number")
		if number < 0.01 {
			return false, "充值金额最低为0.01元"
		}

		//检查付款的账号是否存在
		if _ok := this.member.FindByNickname(this.GetString("account")); !_ok { //如果存在则继续
			return false, "充值帐号不存在"
		}

		//付款平台
		if this.GetString("plantform") != "alipay" {
			return false, "付款平台错误"
		}

		//付款类型
		if this.GetString("type") != "web" {
			return false, "付款类型错误"
		}

		//实例化账单
		order := &models.Order{}
		order.ToName = this.GetString("account")
		order.ToId = this.member.Id
		order.FromName = this.member.Nickname
		order.FromId = this.member.Id
		order.Pay = float32(number)
		order.Type = this.GetString("type")
		order.PayPlantform = this.GetString("plantform")
		order.Description = "网站充值"

		//插入新账单
		order.InsertOrder()

		//获取提交表单
		var form string
		switch order.PayPlantform {
		case "alipay": //支付宝
			form = alipay.CreateAlipaySign(order.Id.Hex(), float32(number), order.ToName, "我酷游戏-充值 "+strconv.FormatFloat(float64(order.Pay), 'f', 2, 32)+" 元")
		}

		//设置操作9秒内不能再次操作的缓存
		models.Redis.Setex(this.member.Id.Hex()+"recharge", 9, []byte("1"))

		return true, form
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 账户充值-post 对已有的未完成账单进行充值 */
func (this *UserController) RechargeOrder() {
	ok, data := func() (bool, interface{}) {
		//根据id查询账单
		order := &models.Order{}
		if !bson.IsObjectIdHex(this.GetString("orderid")) {
			return false, "账单不存在"
		}

		order.Id = bson.ObjectIdHex(this.GetString("orderid"))
		if ok := order.FindOne(); ok != true {
			return false, "账单不存在"
		}

		if order.FromId != this.member.Id {
			return false, "不是您的账单"
		}

		if order.Success == true {
			return false, "该账单已支付完成"
		}

		if time.Now().Sub(order.Time).Hours() > 2 {
			return false, "该账单已失效"
		}

		//获取支付宝即时到帐的自动提交表单
		form := alipay.CreateAlipaySign(order.Id.Hex(), order.Pay, order.ToName, "我酷游戏-充值 "+strconv.FormatFloat(float64(order.Pay), 'f', 2, 32)+" 元")

		return true, form
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 修改绑定邮箱 */
func (this *UserController) Email() {
	ok, data := func() (bool, interface{}) {
		if _, err := models.Redis.Get(this.member.Id.Hex() + "email"); err == nil {
			return false, "您操作过于频繁，3秒后再试"
		}

		if id := this.member.EmailExist(this.GetString("email")); id != "" { //邮箱存在
			return false, "此邮箱已被使用"
		}

		//生成参数
		_, urlParams := this.member.CreateSign(time.Now().Unix()+3600, "email", this.GetString("email"))
		//发送验证邮件
		SendEmail([]string{this.GetString("email")}, this.member.Nickname+"：您正在使用重置邮箱服务，请尽快完成操作"+time.Now().String(), `
					您好：`+this.member.Nickname+`<br><br>
					（请在一小时内完成）您需要点击以下链接来重置您的邮箱（<b>注意：重置邮箱后，原邮箱将与您的账户解绑！</b>）：<br><br>
					http://`+beego.AppConfig.String("httpWebSite")+`/auth`+urlParams)
		this.Data["json"] = []string{"true", ""}

		//设置操作9秒内不能再次操作的缓存
		models.Redis.Setex(this.member.Id.Hex()+"email", 2, []byte("1"))

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 获取第三方平台绑定状况列表 */
func (this *UserController) OauthList() {
	//查询该用户绑定的平台
	party := &models.Party{}
	partys := party.FindBinds(this.member.Id)

	this.Data["json"] = map[string]interface{}{
		"ok":   true,
		"data": partys,
	}
	this.ServeJson()
}

/* 第三方平台绑定 */
func (this *UserController) Oauth() {
	ok, data := func() (bool, interface{}) {
		//检测token是否合法
		if ok := BaiduSocialCheck(this.GetString("token"), this.GetString("id")); !ok {
			return false, "token不合法"
		}

		var plantform string
		switch this.GetString("type") {
		case "baidu":
			plantform = "百度账号"
		case "qqdenglu":
			plantform = "qq账号"
		case "sinaweibo":
			plantform = "新浪微博账号"
		case "renren":
			plantform = "人人网账号"
		case "qqweibo":
			plantform = "腾讯微博"
		case "kaixin":
			plantform = "开心网"
		}

		party := &models.Party{}
		//查询是否存在
		if ok := party.IdExist(this.GetString("id")); ok { //已存在，是更新授权操作
			party.RefreshAuthor(this.GetString("id"), this.GetString("type"), this.GetString("token"), this.GetString("expire"))
			AddMessage(this.member.Id.Hex(), "system", "refreshOauth", plantform+"&"+party.ExpiresTime.String())
		} else { //不存在，新增第三方关联
			party.Insert(this.GetString("id"), this.GetString("type"), this.member.Id.Hex(), this.GetString("token"), this.GetString("nickname"), this.GetString("image"), this.GetString("expire"))
			AddMessage(this.member.Id.Hex(), "system", "addOauth", plantform)
		}

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 修改用户头像post */
func (this *UserController) ChangeImage() {
	this.member.Image = this.GetString("image")
	this.member.Update(bson.M{"$set": bson.M{"i": this.member.Image}})
	AddMessage(this.member.Id.Hex(), "system", "updateImage", this.member.Image)

	this.Data["json"] = map[string]interface{}{
		"ok":   true,
		"data": nil,
	}
	this.ServeJson()
}

/* 网站管理 账号管理 */
func (this *UserController) Member() {
	if this.member.Type != 0 {
		this.StopRun()
	}
	// restful操作接口
	member := &models.Member{}
	models.Restful(member, &this.Controller)
}

/* 查询职位列表 */
func (this *UserController) Jobs() {
	ok, data := func() (bool, interface{}) {
		if this.member.Type != 0 {
			return false, "没有权限"
		}

		job := &models.Job{}
		result := job.FindAll()

		return true, result
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 网站管理 职位管理 */
func (this *UserController) Job() {
	if this.member.Type != 0 {
		this.StopRun()
	}
	// restful操作接口
	job := &models.Job{}
	models.Restful(job, &this.Controller)
}

/*
*
*
*
*
*
 */

/* 账户基本信息 */
func (this *UserController) AccountBase() {
	//查询账户类型
	var typeName string
	if this.member.Type <= 3 {
		switch this.member.Type {
		case 0:
			typeName = "董事长"
		case 1:
			typeName = "普通会员"
		case 2:
			typeName = "高级会员"
		case 3:
			typeName = "白金会员"
		}
	} else {
		job := &models.Job{}
		//查询某个职位
		job.FindOne(this.member.Type)
		typeName = job.Name
	}
	this.Data["typeName"] = typeName
	this.Data["member"] = this.member
	this.TplNames = "user/account/base.html"
	this.Render()
}

/* 领取工资 */
func (this *UserController) JobMoney() {
	//除了会员都可以访问
	if this.member.Type > 0 && this.member.Type < 4 {
		this.StopRun()
	}
	//查询自己分类下的属性
	job := &models.Job{}
	job.FindOne(this.member.Type)
	this.Data["job"] = job
	this.TplNames = "user/job/money.html"
	this.Render()
}

/* 我要升职 */
func (this *UserController) JobPromotion() {
	//除了会员都可以访问
	if this.member.Type > 0 && this.member.Type < 4 {
		this.StopRun()
	}
	this.TplNames = "user/job/promotion.html"
	this.Render()
}

/* 薪资管理 */
func (this *UserController) JobSalary() {
	//0
	if this.member.Type != 0 {
		this.StopRun()
	}
	this.TplNames = "user/job/salary.html"
	this.Render()
}

/* 用户登陆，为用户建立session */
func (this *UserController) UserLogin(contro *beego.Controller) {
	//用户登陆次数加1
	this.member.LogTime++
	//保存用户信息
	this.member.Save()
	//建立session
	contro.SetSession("WOKUID", this.member.Id.Hex())
}

/* 查询用户具体信息 */
func (this *UserController) getUserDetail() interface{} {
	//查询用户的日传图总流量
	job := &models.Job{}
	job.FindOne(this.member.Type)

	//计算用户类型名称
	var typeName string
	switch this.member.Type {
	case 0:
		typeName = "董事长"
	case 1:
		typeName = "会员"
	case 2:
		typeName = "高级会员"
	case 3:
		typeName = "钻石会员"
	default:
		typeName = job.Name
	}

	return map[string]interface{}{
		"nickName":      this.member.Nickname,
		"id":            this.member.Id,
		"email":         this.member.Email,
		"image":         this.member.Image,
		"money":         this.member.Money,
		"uploadSize":    this.member.UploadSize,
		"maxUploadSize": job.UploadSize,
		"free":          this.member.MonthFree,
		"messageNumber": this.member.MessageNumber,
		"lastTime":      this.member.LastTime, //最后登录时间
		"token":         this.member.Token,
		"type":          this.member.Type,
		"typeName":      typeName,
		"power":         this.member.Power,
	}
}

/* 清空消息数量 */
func (this *UserController) ClearMessage() {
	this.member.MessageNumber = 0
	this.member.Save()
}

func (this *UserController) Finish() {
	//保存更新后数据
	this.member.UpdateFinish()
}
