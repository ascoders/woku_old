package controllers

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"woku/models"

	"github.com/ascoders/alipay"
	"github.com/ascoders/upyun"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils"
	"gopkg.in/mgo.v2/bson"
)

/* ----------此控制器定义全部第三方API操作---------- */

type ApiController struct {
	beego.Controller
}

var (
	ApiLog *logs.BeeLogger //打印日志
	mail   *utils.Email    //邮件
	u      *upyun.UpYun    //又拍云
)

func init() {
	//如果没有日志目录则创建日志目录
	_, err := os.Open("log")
	if err != nil && os.IsNotExist(err) {
		os.Mkdir("log", 0777)
	}
	//初始化日志
	ApiLog = logs.NewLogger(10000)
	ApiLog.SetLogger("file", `{"filename":"log/api.log"}`)
	//初始化邮箱
	mail = utils.NewEMail(`{"username":"` + beego.AppConfig.String("EmailAccount") + `","password":"` + beego.AppConfig.String("EmailPassword") + `","host":"` + beego.AppConfig.String("EmailSmtp") + `","port":` + beego.AppConfig.String("EmailPort") + `}`)
	//初始化又拍云
	u = upyun.NewUpYun("wokugame", beego.AppConfig.String("UpyunAccount"), beego.AppConfig.String("UpyunPassword"))
	u.SetApiDomain("v0.api.upyun.com")
}

/* 重写CheckXsrfCookie方法 */
func (this *ApiController) CheckXsrfCookie() bool {
	return true
}

/* ---------------- 又拍云 ---------------- */

/* 获取又拍云表单上传token */
func GetUpyunUploadToken() {
	type Options struct {
		Bucket      string `json:"bucket"`             //空间名
		Expiration  int64  `json:"expiration"`         //授权过期时间
		Path        string `json:"save-key"`           //保存路径
		Type        string `json:"allow-file-type"`    //文件类型限制
		WidthRange  string `json:"image-width-range"`  //宽度范围
		HeightRange string `json:"image-height-range"` //高度范围
	}
	//实例化上传数组信息
	options := &Options{}
	options.Bucket = "wokugame"
	options.Expiration = time.Now().Unix() + 60
	options.Path = "/{year}/{mon}/{day}/{random32}"
	options.Type = "jpg,gif,png"
	options.WidthRange = "0,1024"
	options.HeightRange = "0,1024"
	//json_encode编码
	encode, _ := json.Marshal(options)
	//base64编码得出policy
	policy := base64.StdEncoding.EncodeToString(encode)
	//计算签名
	m := md5.New()
	m.Write([]byte(policy + "&" + beego.AppConfig.String("UpyunFormToken")))
	//signature := hex.EncodeToString(m.Sum(nil))
	//result := []string{policy, signature}
}

/* 根据条件保存上传文件 */
func (this *ApiController) SaveFile(contro *beego.Controller, size int64, suf []string) (bool, string, int64) {
	//获取文件信息
	f, h, err := contro.GetFile("file")
	if err != nil {
		UserLog.Error("文件上传失败：", err)
		f.Close()
		return false, "", 0
	}
	f.Close()
	//保存文件到临时目录
	contro.SaveToFile("file", "static/tmp/"+h.Filename)
	// 打开文件
	fh, err := os.Open("static/tmp/" + h.Filename)
	if err != nil {
		return false, "", 0
	}
	//获取文件具体信息
	stat, _ := fh.Stat()
	if stat.Size() > size { //文件大小不合格，删除文件
		//关闭文件
		fh.Close()
		os.Remove("static/tmp/" + h.Filename)
		return false, "", 0
	}
	nameSplit := strings.Split(h.Filename, ".")
	suffix := nameSplit[len(nameSplit)-1]
	ok := false
	//后缀不合格则删除文件
	for _, v := range suf {
		if suffix == v { //符合其中一个后缀
			ok = true
		}
	}
	if !ok {
		//关闭文件
		fh.Close()
		os.Remove("static/tmp/" + h.Filename)
		return false, "", 0
	}
	//关闭文件
	fh.Close()
	return true, "static/tmp/" + h.Filename, stat.Size() //返回文件路径、大小
}

/* 异步上传文件，服务器接收文件后上传到又拍云，删除本地文件
 * @params static bool 静态，false时会根据年月日存在不同子文件夹内
 */
func (this *ApiController) UploadFile(path string, remoteDir string, static bool) (bool, string) {
	// 打开文件
	fh, err := os.Open(path)
	// 设置待上传文件的 Content-MD5 值（如又拍云服务端收到的文件MD5值与用户设置的不一致，将回报 406 Not Acceptable 错误）
	u.SetContentMD5(upyun.FileMd5(path))
	//定义目录：年份_月份/日/随机文件名.后缀
	var dirPath string
	if beego.RunMode == "prod" {
		dirPath = "/" + remoteDir + "/"
	} else {
		dirPath = "/test/" + remoteDir + "/"
	}
	dir := dirPath
	if !static { //不是静态存储，加上年月日父级文件夹
		dir += strconv.Itoa(time.Now().Year()) + "_" + time.Now().Month().String() + "/" + strconv.Itoa(time.Now().Day()) + "/"
	}
	//随机文件名，随机时间种子
	rands := rand.New(rand.NewSource(time.Now().UnixNano()))
	randName := rands.Int31()
	//获取文件后缀
	nameSplit := strings.Split(path, ".")
	suffix := nameSplit[len(nameSplit)-1]
	//上传文件
	err = u.WriteFile(dir+strconv.Itoa(int(randName))+"."+suffix, fh, true)
	//如果有错误，记录错误
	if err != nil {
		ApiLog.Trace("上传图片到又拍云: %v\n", err)
	}
	//关闭文件
	fh.Close()
	//上传完毕后删除文件
	os.Remove(path)
	if err != nil { //上传又拍云失败
		return false, ""
	} else { //上传成功
		return true, dir + strconv.Itoa(int(randName)) + "." + suffix
	}
}

/* 删除本地文件 */
func DeleteFile(path string) {
	os.Remove(path)
}

/* 删除又拍云文件 */
func DeleteOriginFile(path string) bool {
	err := u.DeleteFile(path)
	if err != nil {
		return false
	} else {
		return true
	}
}

/* ---------------- 支付宝 ---------------- */

type AlipayParameters struct {
	InputCharset string `json:"_input_charset"` //网站编码
	Body         string `json:"body"`           //订单描述
	NotifyUrl    string `json:"notify_url"`     //异步通知页面
	OutTradeNo   string `json:"out_trade_no"`   //订单唯一id
	Partner      string `json:"partner"`        //合作者身份ID
	PaymentType  uint8  `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string `json:"return_url"`     //回调url
	SellerEmail  string `json:"seller_email"`   //卖家支付宝邮箱
	Service      string `json:"service"`        //接口名称
	Subject      string `json:"subject"`        //商品名称
	TotalFee     int    `json:"total_fee"`      //总价
	Sign         string `json:"sign"`           //签名，生成签名时忽略
	SignType     string `json:"sign_type"`      //签名类型，生成签名时忽略
}

/* 按照支付宝规则生成sign */
func AlipaySign(param interface{}) string {
	//解析为字节数组
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return ""
	}
	//重组字符串
	var sign string
	oldString := string(paramBytes)
	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)
	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)
	paramArray := strings.Split(oldString, ",")
	for _, v := range paramArray {
		detail := strings.SplitN(v, ":", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}
	}
	//追加密钥
	sign += beego.AppConfig.String("AlipayKey")
	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}

/* 主动向支付宝查询某个订单是否交易成功 :todo */
func (this *ApiController) CheckAlipayOrder() bool {
	type PostParam struct {
		InputCharset string `json:"_input_charset"` //网站编码
		OutTradeNo   string `json:"out_trade_no"`   //订单唯一id
		Partner      string `json:"partner"`        //合作者身份ID
		Service      string `json:"service"`        //接口名称
		Sign         string `json:"sign"`           //签名，生成签名时忽略
		SignType     string `json:"sign_type"`      //签名类型，生成签名时忽略
	}
	//实例化参数
	param := &PostParam{}
	param.InputCharset = "utf-8"
	param.OutTradeNo = "535a2046a6e10b0838000007"
	param.Partner = beego.AppConfig.String("AlipayPartner")
	param.Service = "single_trade_query"
	//生成签名
	//sign := AlipaySign(param)
	//生成url
	//url := "https://mapi.alipay.com/gateway.do?" + tempSign + "&sign=" + sign + "&sign_type=MD5"
	//httplib.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//req := httplib.Get("https://mapi.alipay.com/gateway.do?service=single_trade_query&sign=d8ed9f015214e7cd59bfadb6c945a87b&trade_no=2010121502730740&partner=2088001159940003&out_trade_no=2109095506028810&sign_type=MD5")
	//fmt.Println(url)
	return false
}

/* 接收支付宝同步跳转的页面 */
func (this *ApiController) AlipayReturn() {
	status, orderId, buyerEmail, tradeNo := alipay.AlipayReturn(&this.Controller)
	if status == 1 { //付款成功，处理订单
		//查询订单
		order := &models.Order{}
		order.Id = bson.ObjectIdHex(orderId)
		if ok := HandleOrder(buyerEmail, tradeNo, order); ok { //充值成功
			this.Data["success"] = true
			this.Data["account"] = order.ToName
			this.Data["gain"] = order.Pay
		} else {
			this.Data["success"] = false
		}
	} else {
		this.Data["success"] = false
	}
	this.TplNames = "api/alipayReturn.html"
	this.Render()
}

/* 被动接收支付宝异步通知的页面 */
func (this *ApiController) AlipayNotify() {
	status, orderId, buyerEmail, TradeNo := alipay.AlipayNotify(&this.Controller)
	if status == 1 { //付款成功，处理订单
		//查询订单
		order := &models.Order{}
		order.Id = bson.ObjectIdHex(orderId)
		HandleOrder(buyerEmail, TradeNo, order)
	}
}

/* ---------------- 支付宝手机版 ---------------- */

/* 被动接收支付宝手机网站支付页面跳转的页面 */
func (this *ApiController) AlipayMobileReturn() {
	//列举全部传参
	type Params struct {
		OutTradeNo   string `form:"out_trade_no" json:"out_trade_no"`   //在网站中唯一id
		RequestToken string `form:"request_token" json:"request_token"` //授权令牌
		Result       string `form:"result" json:"result"`               //支付结果 success
		TradeNo      string `form:"trade_no" json:"trade_no"`           //支付宝交易号
		Sign         string `form:"sign" json:"sign"`                   //签名
	}
	//实例化参数
	param := &Params{}
	//解析表单内容
	if err := this.ParseForm(param); err != nil {
		this.StopRun()
	}
	//如果最基本的网站交易号为空，则跳转
	if param.OutTradeNo == "" { //不存在交易号
		this.StopRun()
	} else {
		//生成签名
		sign := AlipaySign(param)
		//对比签名是否相同
		if sign == param.Sign { //只有相同才说明该订单成功了
			//判断订单是否已完成
			if param.Result == "success" { //交易成功
				//查询订单
				order := &models.Order{}
				order.Id = bson.ObjectIdHex(param.OutTradeNo)
				//处理账单
				if ok := HandleOrder("", param.TradeNo, order); ok { //充值成功

				} else { //已经处理过了

				}
				//查询游戏
				game := &models.Game{}
				if ok := game.FindPath(order.Game); !ok { //游戏不存在
					this.StopRun()
				}
				this.Data["gameName"] = game.Name
				this.Data["gain"] = order.Gain
				this.Data["pay"] = order.Pay
				this.TplNames = "api/mobilePayReturn.html"
				this.Render()
			} else {
				this.StopRun()
			}
		} else {
			this.StopRun()
		}
	}
}

/* 生成支付宝手机支付request_token授权令牌，并调用createAlipayMobileSign构造签名并生成支付请求表单
 * @params ObjectId 订单唯一id
 * @params int 价格
 * @params int 获得代金券的数量
 * @params string 充值账户的名称
 */
func CreateAlipayMobileToken(orderId bson.ObjectId, fee float32, account string, description string) string {
	type AlipayMobileParameters struct {
		InputCharset string `json:"_input_charset"` //网站编码
		Format       string `json:"format"`         //订单描述
		Partner      string `json:"partner"`        //合作者身份ID
		ReqData      string `json:"req_data"`       //支付信息
		ReqId        string `json:"req_id"`         //不重复id
		SecId        string `json:"sec_id"`         //签名加密方式
		Service      string `json:"service"`        //请求api类型
		V            string `json:"v"`              //接口版本号
		Sign         string `json:"sign"`           //签名，生成签名时忽略
	}
	//金额变成字符串
	feeString := strconv.FormatFloat(float64(fee), 'f', 2, 32)
	//实例化参数
	param := &AlipayMobileParameters{}
	param.InputCharset = "utf-8"
	param.Format = "xml"
	param.Partner = beego.AppConfig.String("AlipayPartner")
	param.ReqData = "<direct_trade_create_req><call_back_url>http://www.wokugame.com/api/alipaymobilereturn.html</call_back_url><notify_url>http://www.wokugame.com/api/alipaynotify</notify_url><out_trade_no>" + orderId.Hex() + "</out_trade_no><out_user>" + account + "</out_user><seller_account_name>576625322@qq.com</seller_account_name><subject> " + description + "</subject><total_fee>" + feeString + "</total_fee></direct_trade_create_req>"
	param.ReqId = orderId.Hex()
	param.SecId = "MD5"
	param.Service = "alipay.wap.trade.create.direct"
	param.V = "2.0"
	//生成签名
	sign := AlipaySign(param)
	//追加参数
	param.Sign = sign
	//实例化url参数
	urls := &url.Values{}
	urls.Add("_input_charset", param.InputCharset)
	urls.Add("format", param.Format)
	urls.Add("partner", param.Partner)
	urls.Add("req_data", param.ReqData)
	urls.Add("req_id", param.ReqId)
	urls.Add("sec_id", param.SecId)
	urls.Add("service", param.Service)
	urls.Add("v", param.V)
	urls.Add("sign", param.Sign)
	//生成请求参数
	urlParam := urls.Encode()
	//get访问支付宝网管
	req := httplib.Post("http://wappaygw.alipay.com/service/rest.htm?" + urlParam)
	//获取支付宝返回信息
	str, _ := req.String()
	//获取res_data信息
	var res_data string
	strArray := strings.Split(str, "&")
	for _, v := range strArray {
		if strings.Index(v, "res_data") != -1 {
			res_data = strings.Split(v, "=")[1]
			break
		}
	}
	res_data, _ = url.QueryUnescape(res_data)
	//获取<request_token></request_token>之间的request_token
	re, _ := regexp.Compile("\\<request_token[\\S\\s]+?\\</request_token>")
	rt := re.FindAllString(res_data, 1)
	request_token := strings.Replace(rt[0], "<request_token>", "", -1)
	request_token = strings.Replace(request_token, "</request_token>", "", -1)
	//调用createAlipayMobileSign
	form := createAlipayMobileSign(request_token)
	return form
}

/* 生成手机支付签名并返回构造表单 */
func createAlipayMobileSign(RequestToken string) string {
	type AlipayMobileParameters struct {
		InputCharset string `json:"_input_charset"` //网站编码
		Format       string `json:"format"`         //订单描述
		Partner      string `json:"partner"`        //合作者身份ID
		ReqData      string `json:"req_data"`       //支付信息
		SecId        string `json:"sec_id"`         //签名加密方式
		Service      string `json:"service"`        //请求api类型
		V            string `json:"v"`              //接口版本号
		Sign         string `json:"sign"`           //签名，生成签名时忽略
	}
	//实例化参数
	param := &AlipayMobileParameters{}
	param.InputCharset = "utf-8"
	param.Format = "xml"
	param.Partner = beego.AppConfig.String("AlipayPartner")
	param.ReqData = "<auth_and_execute_req><request_token>" + RequestToken + "</request_token></auth_and_execute_req>"
	param.SecId = "MD5"
	param.Service = "alipay.wap.auth.authAndExecute"
	param.V = "2.0"
	//生成签名
	sign := AlipaySign(param)
	//追加参数
	param.Sign = sign
	//生成自动提交表单
	form := `
		<form id="alipaysubmit" name="alipaysubmit" action="http://wappaygw.alipay.com/service/rest.htm?_input_charset=utf-8" method="get" style='display:none;'>
			<input type="hidden" name="_input_charset" value="` + param.InputCharset + `">
			<input type="hidden" name="format" value="` + param.Format + `">
			<input type="hidden" name="partner" value="` + param.Partner + `">
			<input type="hidden" name="req_data" value="` + param.ReqData + `">
			<input type="hidden" name="sec_id" value="` + param.SecId + `">
			<input type="hidden" name="service" value="` + param.Service + `">
			<input type="hidden" name="v" value="` + param.V + `">
			<input type="hidden" name="sign" value="` + param.Sign + `">
		</form>
		<script>document.forms['alipaysubmit'].submit();</script>
	`
	return form
}

/* 手机支付统一跳转页面 */
func (this *ApiController) MobilePay() {
	/*
		//获取过期时间的秒数
		expire, _ := this.GetInt("expire")
		//获取付款金额
		pay, _ := this.GetFloat("pay")
		//待计算的实际到帐金额
		var gain int
		//帐户昵称
		var nickname string
		//账户objectid
		var objectid string
		//过期则退出
		if !time.Unix(expire, 0).After(time.Now()) {
			this.StopRun()
		}
		//实例化账单
		order := &models.Order{}
		order.ToId = bson.ObjectIdHex(objectid)
		order.ToName = nickname
		order.FromId = bson.ObjectIdHex(objectid)
		order.FromName = nickname
		order.Pay = float32(pay)
		order.Type = this.GetString("game")
		//插入新账单
		order.InsertOrder()
		//调用手机网站支付api
		form := CreateAlipayMobileToken(order.Id, int(pay), gain, objectid)
		this.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
		this.Ctx.WriteString(form)
	*/
}

/* ---------------- 阿里云企业邮箱 ---------------- */

/* 发送邮件 */
func SendEmail(address []string, subject string, html string) {
	mail.To = address
	mail.From = beego.AppConfig.String("EmailAccount")
	mail.Subject = subject
	mail.Text = ""
	mail.HTML = `
			<div style="border-bottom:3px solid #d9d9d9; background:url(http://www.wokugame.com/static/img/email_bg.gif) repeat-x 0 1px;">
				<div style="border:1px solid #c8cfda; padding:40px;">
					` + html + `
					<p>&nbsp;</p>
					<div>我酷游戏团队 祝您游戏愉快</div>
					<div>Powered by wokugame</div>
					<img src="http://www.wokugame.com/static/img/logo.png">
					</div>
				</div>
			</div>
			`
	err := mail.Send()
	if err != nil { //邮件未发送成功，记录错误日志
		ApiLog.Error("邮件发送失败：", err)
	}
}

/* ---------------- 百度开放云 ---------------- */

/* 判断token对应的id是否合法 */
func BaiduSocialCheck(token string, id string) bool {
	var r = struct {
		Id uint `json:"social_uid"`
	}{}
	err := httplib.Get("https://openapi.baidu.com/social/api/2.0/user/info").Param("access_token", token).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).ToJson(&r)
	if err != nil { //请求发送失败
		return false
	}
	if strconv.Itoa(int(r.Id)) != id { //id不符
		return false
	}
	return true
}

/* ---------------- 对外统一api ---------------- */
func (this *ApiController) Api() {
	switch this.GetString("type") {
	case "mobile_pay": //付款
		//@params id string 合作者账号id
		//@params expire string 过期时间 例：3600
		//@params pay string 付款金额 例：0.01
		//@params reqid string 充值账户唯一id
		//@params name string 付款游戏/应用名称 例：看准颜色
		//@params gain string 获取道具 例：5宝石
		//@params notify string 回调地址 例：http://www.example.com/notify
		//@params extend string 拓展参数
		//@params sign string 签名
		if !this.Ctx.Input.IsAjax() {
			//查找游戏
			game := &models.Game{}
			if ok := game.FindPath(this.GetString("game")); !ok { //游戏不存在
				this.Abort("404")
				return
			}
			this.Data["gameName"] = game.Name
			this.Data["gain"] = this.GetString("gain")
			this.TplNames = "api/mobilePay.html"
			this.Render()
		} else {
			defer this.ServeJson()
			//支付平台
			if this.GetString("plantform") != "alipay" {
				this.Data["json"] = "支付平台不在范围内"
				return
			}
			//查找游戏
			game := &models.Game{}
			if ok := game.FindPath(this.GetString("game")); !ok { //游戏不存在
				this.Data["json"] = "该游戏尚未注册"
				return
			}
			//对比签名是否正确
			member := &models.Member{}
			if ok := member.FindOne(this.GetString("id")); !ok { //合作账号不存在
				this.Data["json"] = "合作账号不存在"
				return
			}
			//创建签名
			expire, err := this.GetInt("expire")
			if err != nil { //过期时间格式不对
				this.Data["json"] = "时间格式错误"
				return
			}
			pay, err := this.GetFloat("pay")
			if err != nil || pay < 0.01 {
				this.Data["json"] = "付款金额错误"
				return
			}
			inputSign := this.GetString("sign")
			plantform := this.GetString("plantform")
			this.Input().Del("sign")                          //删除sign字段
			this.Input().Del("plantform")                     //删除交易平台字段
			sign := MD5(this.Input().Encode() + member.Token) //排序后encode字段加上密钥md5加密生成token
			if sign != inputSign {                            //签名不符合
				this.Data["json"] = "签名不符"
				return
			}
			if !time.Unix(int64(expire), 0).After(time.Now()) {
				this.Data["json"] = "该订单已失效"
				return
			}
			//实例化账单
			order := &models.Order{}
			order.ToId = member.Id
			order.ToName = member.Nickname
			order.FromId = member.Id
			order.FromName = member.Nickname
			order.Pay = float32(pay)
			order.PayPlantform = plantform
			order.Description = game.Name + " " + this.GetString("reqid") + " 购买 " + this.GetString("gain")
			order.Type = "mobile"
			order.Notify = this.GetString("notify")
			order.Reqid = this.GetString("reqid")
			order.Game = this.GetString("game")
			order.Gain = this.GetString("gain")
			order.Extend = this.GetString("extend")
			//插入新账单
			order.InsertOrder()
			//调用手机网站支付api
			var form string
			switch order.PayPlantform {
			case "alipay": //支付宝
				form = CreateAlipayMobileToken(order.Id, float32(pay), member.Id.Hex(), game.Name+":"+this.GetString("gain"))
				this.Data["json"] = form
			}
		}
	default:
		this.Abort("404")
		return
	}
}
