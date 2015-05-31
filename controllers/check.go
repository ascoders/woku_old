package controllers

import (
	"github.com/dchest/captcha"
	"regexp"
	"strings"
	"time"
	"woku/models"

	"github.com/astaxie/beego"
)

type CheckController struct {
	beego.Controller
}

/* 登陆 */
func (this *CheckController) Login() {
	ok, data := func() (bool, interface{}) {
		if len([]rune(this.GetString("account"))) < 3 || len([]rune(this.GetString("account"))) > 30 {
			return false, "账号长度为3~30"
		}

		if ok, err := regexp.MatchString("[0-9a-zA-Z_-]{6,30}", this.GetString("password")); !ok || err != nil {
			return false, "密码长度为6~30的字母或数字"
		}

		//判断登陆
		user := &UserController{}
		info, statu := user.member.CheckPass(this.GetString("account"), this.GetString("password"))

		switch statu {
		case 1:
			//用户登陆
			user.UserLogin(&this.Controller)
			//查询用户信息
			return true, user.getUserDetail()
		case -1:
			return false, "密码错误"
		case -2:
			return false, "账号不存在"
		case -3:
			return false, "帐号已锁定，" + info + "后恢复"
		case -4:
			return false, "账号暂时锁定"
		}
		return false, "未知错误"
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 根据token自动登陆 */
func (this *CheckController) Auth() {
	ok, data := func() (bool, interface{}) {
		//实例化用户
		user := &UserController{}

		//获取过期时间的秒数
		expire, _ := this.GetInt("expire")
		if _ok := user.member.FindOne(this.GetString("id")); !_ok && this.GetString("type") != "createAccount" { //账户不存在
			return false, "账号不存在"
		}

		if time.Unix(int64(expire), 0).Before(time.Now()) { // 时间过期
			return false, "该请求已过期"
		}

		sign, _ := user.member.CreateSign(int64(expire), this.GetString("type"), this.GetString("extend")) //创建服务端签名
		if this.GetString("sign") != sign {
			return false, "签名错误"
		}

		if this.GetString("type") != "createAccount" {
			//用户登陆
			user.UserLogin(&this.Controller)
		}

		//根据类型执行操作
		switch this.GetString("type") {
		case "email":
			user.member.Email = this.GetString("extend")
			user.member.Save()
		case "createAccount": //注册账号
			// 邮箱 昵称 密码
			params := strings.Split(this.GetString("extend"), "|")

			//填充用户信息
			user := &UserController{}
			user.member.Email = params[0]
			user.member.Nickname = params[1]
			user.member.Password = params[2]
			user.member.Type = 1
			//检测账户是否已存在
			if objectid := user.member.EmailExist(user.member.Email); objectid != "" {
				return false, "账户已存在"
			}

			//检测账户用户名是否已存在
			if objectid := user.member.NicknameExist(user.member.Nickname); objectid != "" {
				return false, "用户名已被使用"
			}

			//注册用户
			user.member.Insert()

			//用户登陆
			user.UserLogin(&this.Controller)

			//提示用户可以绑定第三方平台账号登陆
			AddMessage(user.member.Id.Hex(), "system", "bindOauth", "")

			//查询用户信息
			return true, user.getUserDetail()
		}

		//查询用户信息
		return true, user.getUserDetail()
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 注册 */
func (this *CheckController) Register() {
	ok, data := func() (bool, interface{}) {
		if len([]rune(this.GetString("email"))) > 100 {
			return false, "邮箱过长"
		}

		if strings.Contains(this.GetString("email"), "|") {
			return false, "邮箱不能含有“|”符号"
		}

		if len([]rune(this.GetString("nickname"))) < 3 || len([]rune(this.GetString("nickname"))) > 20 {
			return false, "昵称长度为3-20"
		}

		if strings.Contains(this.GetString("nickname"), "|") {
			return false, "昵称不能含有“|”符号"
		}

		if len([]rune(this.GetString("password"))) < 6 || len([]rune(this.GetString("password"))) > 30 {
			return false, "密码长度为6-30"
		}

		if strings.Contains(this.GetString("password"), "|") {
			return false, "密码不能含有“|”符号"
		}

		if this.GetSession("WOKUID") != nil { //已有session则退出
			return false, "已登录"
		}

		//验证码校验
		cap_id, cap_value := this.GetString("capid"), this.GetString("cap")
		if ok := captcha.VerifyString(cap_id, cap_value); !ok {
			return false, "验证码错误"
		}

		//数据赋值
		user := &UserController{}
		user.member.Email = this.GetString("email")
		user.member.Nickname = this.GetString("nickname")
		user.member.Password = this.GetString("password")
		user.member.Type = 1 //普通会员

		//检测邮箱是否已存在
		if objectid := user.member.EmailExist(this.GetString("email")); objectid != "" {
			return false, "您的账号已存在，可以直接登录"
		}

		//检测昵称是否存在
		if objectid := user.member.NicknameExist(this.GetString("nickname")); objectid != "" {
			return false, "昵称已被使用"
		}

		//生成参数
		_, urlParams := user.member.CreateSign(time.Now().Unix()+3600, "createAccount", this.GetString("email")+"|"+this.GetString("nickname")+"|"+this.GetString("password"))

		//发送验证邮件
		SendEmail([]string{user.member.Email}, user.member.Nickname+"：您的我酷账号申请成功，请点击链接激活！"+time.Now().String(), `
		您好：`+user.member.Nickname+`<br><br>
		（请在一小时内完成）您需要点击以下链接来激活您的我酷账户：<br><br>
		http://`+beego.AppConfig.String("httpWebSite")+`/auth`+urlParams)

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 登出 */
func (this *CheckController) Signout() {
	this.DelSession("WOKUID")
	this.Data["json"] = map[string]interface{}{
		"ok": true,
	}
	this.ServeJson()
}

/* 获取登陆用户信息 */
func (this *CheckController) CurrentUser() {
	ok, data := func() (bool, interface{}) {
		//是否存在session
		if this.GetSession("WOKUID") == nil || this.GetSession("WOKUID").(string) == "" {
			return false, "未登录"
		}

		//查找用户
		user := UserController{}
		if ok := user.member.FindOne(this.GetSession("WOKUID").(string)); !ok {
			return false, "不存在的用户"
		}

		//获取用户信息
		return true, user.getUserDetail()
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}

	this.ServeJson()
}

/* 刷新验证码 */
func (this *CheckController) FreshCap() {
	this.Data["json"] = map[string]interface{}{
		"ok":   true,
		"data": captcha.NewLen(6),
	}
	this.ServeJson()
}

/* 社交化平台查询是否有账号，若有自动登陆 */
func (this *CheckController) HasOauth() {
	ok, data := func() (bool, interface{}) {
		party := &models.Party{}
		//查询账号是否存在
		ok := party.FindMember(this.GetString("id"), this.GetString("type"))
		if !ok { //用户不存在，显示创建用户信息
			return true, -1
		}

		//用户已存在
		if !BaiduSocialCheck(this.GetString("token"), this.GetString("id")) { //通过不验证
			return false, "验证失败"
		}

		//刷新验证权限
		party.RefreshAuthor(this.GetString("id"), this.GetString("type"), this.GetString("token"), this.GetString("expire"))

		//实例化用户
		user := UserController{}
		user.member.FindOne(party.MemberId.Hex()) //查询用户

		//用户登陆
		user.UserLogin(&this.Controller)

		//查询用户信息
		return true, user.getUserDetail()
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}

	this.ServeJson()
}

/* 第三方平台注册用户 */
func (this *CheckController) OauthRegister() {
	ok, data := func() (bool, interface{}) {
		user := UserController{}
		party := &models.Party{}

		//检查昵称长度
		if len([]rune(this.GetString("nickname"))) < 3 || len([]rune(this.GetString("nickname"))) > 20 {
			return false, "昵称长度为3-20"
		}

		if strings.Contains(this.GetString("nickname"), "|") {
			return false, "昵称不能含有“|”符号"
		}

		//查找是否有重复的昵称
		if ok := user.member.NicknameExist(this.GetString("nickname")); ok != "" {
			return false, "昵称重复"
		}

		//查找是否有重复第三方id
		if ok := party.IdExist(this.GetString("id")); ok { //已存在
			return false, "该平台已经注册了账号"
		}

		//检测token是否合法
		if ok := BaiduSocialCheck(this.GetString("token"), this.GetString("id")); !ok {
			return false, "参数非法"
		}

		//插入用户
		user.member.Nickname = this.GetString("nickname") //用户手动输入的昵称，默认是该平台的昵称
		user.member.Image = this.GetString("image")
		user.member.Type = 1
		id := user.member.Insert() //插入成功返回用户id

		//插入第三方关联
		party.Insert(this.GetString("id"), this.GetString("type"), id, this.GetString("token"), this.GetString("nickname"), this.GetString("image"), this.GetString("expire"))

		//提示用户可以绑定邮箱和设置密码
		AddMessage(user.member.Id.Hex(), "system", "bindEmail", "")
		//建议您尽快绑定邮箱、设置密码", "亲爱的用户，您刚刚通过第三方平台注册，但由于大部分服务需要和您邮箱进行交互，建议您立即绑定邮箱！如果同时在账号管理->修改密码操作的话，您下次也可以使用邮箱+密码的方式登录网站。

		//用户登陆
		user.UserLogin(&this.Controller)

		//查询用户信息
		return true, user.getUserDetail()
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}

	this.ServeJson()
}

/*
*
*
*
*
 */

/* 找回密码 */
func (this *CheckController) FindPass() {
	this.TplNames = "check/findpass.html"
	this.Render()
}

/* 找回密码post */
func (this *CheckController) FindPassPost() {
	if IsBusy("findpass", this.Ctx.Input.IP(), 9) { //此ip9秒内不得再操作
		this.Data["json"] = []string{"false", "操作过于频繁"}
	} else {
		//接收type和value两个参数
		switch this.GetString("type") {
		case "email": //邮件找回密码
			//查找这个邮箱用户是否存在
			member := &models.Member{}
			if ok := member.FindByAccount(this.GetString("value")); ok == false { //不存在此email
				this.Data["json"] = []string{"false", "不存在此邮箱地址"}
			} else {
				//生成参数
				_, urlParams := member.CreateSign(time.Now().Unix()+3600, "resetpass", "")
				//发送验证邮件
				SendEmail([]string{member.Email}, member.Nickname+"：您正在使用邮箱找回密码服务，请尽快完成操作"+time.Now().String(), `
					您好：`+member.Nickname+`<br><br>
					（请在一小时内完成）您需要点击以下链接来重置您的密码：<br><br>
					http://`+beego.AppConfig.String("httpWebSite")+`/login.html`+urlParams)
				this.Data["json"] = []string{"true", ""}
				Mc.Put("findpass"+this.Ctx.Input.IP(), 1, 9) //9秒冷却时间
			}
		}
	}
	this.ServeJson()
}
