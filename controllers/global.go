package controllers

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html"
	"math"
	"net/url"
	"os"
	//"path/filepath"
	"github.com/deckarep/golang-set"
	"regexp"
	"strconv"
	"strings"
	"time"
	"woku/models"

	//"code.google.com/p/mahonia"
	//"github.com/PuerkitoBio/goquery"
	"github.com/ascoders/alipay"
	//"github.com/ascoders/html2md"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"gopkg.in/mgo.v2/bson"
)

type GlobalController struct {
	beego.Controller
}

var (
	Bm                cache.Cache     //文件缓存
	Mc                cache.Cache     //内存缓存
	mobileBrowserList []string        //手机信息
	GlobalLog         *logs.BeeLogger //全局日志
)

func init() {
	//初始化手机浏览器识别
	mobileBrowserList = []string{"Android", "iPhone", "iPod", "iPad", "Windows Phone", "MQQBrowser"}

	//初始化支付宝插件
	alipay.AlipayPartner = beego.AppConfig.String("AlipayPartner")
	alipay.AlipayKey = beego.AppConfig.String("AlipayKey")
	alipay.WebReturnUrl = "http://www.wokugame.com/api/alipayreturn.html"
	alipay.WebNotifyUrl = "http://www.wokugame.com/api/alipaynotify"
	alipay.WebSellerEmail = "576625322@qq.com"

	//如果没有日志目录则创建日志目录
	_, err := os.Open("log")
	if err != nil && os.IsNotExist(err) {
		os.Mkdir("log", 0777)
	}

	//初始化日志
	GlobalLog = logs.NewLogger(10000)
	GlobalLog.SetLogger("file", `{"filename":"log/global.log"}`)

	//初始化beego验证提示信息
	validation.MessageTmpls = map[string]string{
		"Required":     "不能为空",
		"Min":          "最小值为 %d",
		"Max":          "最大值为 %d",
		"Range":        "范围为 %d ~ %d",
		"MinSize":      "最小长度为 %d",
		"MaxSize":      "最大长度为 %d",
		"Length":       "长度必须为 %d",
		"Alpha":        "必须为alpha字符",
		"Numeric":      "必须为数字",
		"AlphaNumeric": "必须为alpha字符或数字",
		"Match":        "必须匹配 %s",
		"NoMatch":      "必须不匹配 %s",
		"AlphaDash":    "必须为alpha字符或数字或横杠-_",
		"Email":        "必须为有效的邮箱地址",
		"IP":           "必须为有效的IP地址",
		"Base64":       "必须为有效base64编码",
		"Mobile":       "必须为有效手机号",
		"Tel":          "必须为有效固话",
		"Phone":        "必须为有效固话或手机号",
		"ZipCode":      "必须为有效邮政编码",
	}
}

/* ----------此控制器定义全局操作---------- */

/* 定期执行计划任务 */
func PlanTask() {
	timer := time.NewTicker(3 * time.Hour)
	for {
		select {
		case <-timer.C: //每3小时定时任务
			go func() {
				return //暂不执行

				//实例化对象
				game := &models.Game{}
				member := &models.Member{}
				topic := &models.Topic{}
				reply := &models.Reply{}
				order := &models.Order{}

				/* 清空缓存目录 */
				os.RemoveAll("cache")

				/* 清除成立一个月以上但活跃度不超过10的游戏 */
				bads := game.FindBads()
				GlobalLog.Notice("删除的游戏：", fmt.Sprintln(bads))
				for k, _ := range bads {
					if ok := member.FindOne(bads[k].Manager.Hex()); ok {
						//删除该游戏所有图片 game/{gamePath}/ 下所有图片
						qiniu := &QiniuController{}
						go qiniu.DeleteAll("woku", "game/"+bads[k].Id+"/")

						//删除所有相关帖子
						topic.DeleteGameTopic(bads[k].Id)

						//删除所有相关评论、回复
						reply.DeleteGameReply(bads[k].Id)

						//生成参数
						_, urlParams := member.CreateSign(time.Now().Unix()+3600, "redirect", "/article/53edfa461fc9010589000004.html")
						url := "http://" + beego.AppConfig.String("httpWebSite") + "/login.html" + urlParams
						//邮件通知管理员
						SendEmail([]string{member.Email}, "抱歉，您的板块 "+bads[k].Name+" 使用率过低，已被清除 "+time.Now().String(), "非常抱歉，根据 <a href='"+url+"'>"+url+"</a> 板块清理规范，您的板块 <b>"+bads[k].Name+"</b> 已被移除，您可以重新申请板块并尝试增加使用率！")
						//推送消息
						//AddMessage(member.Id.Hex(), "system", "", "您的板块 "+bads[k].Name+" 使用率过低，已被清除", "非常抱歉，根据 <a href='http://www.wokugame.com/article/53edfa461fc9010589000004.html'>http://www.wokugame.com/article/53edfa461fc9010589000004.html</a> 板块清理规范，您的板块 <b>"+bads[k].Name+"</b> 已被移除，您可以重新申请板块并尝试增加使用率！", "")
						//删除游戏
						bads[k].Delete()
					}
				}

				/* 删除查询游戏列表缓存 */
				models.DeleteCaches("game-find-")
				models.DeleteCaches("game-findcount-")

				/* 删除已失效订单 */
				order.DeleteBad()

				/* 刷新sitemap */
				refreshSitemap()
			}()
		}
	}
}

/* 更新sitemap */
func refreshSitemap() {
	numPerXml := 10000 //每个sitemap里文件数量

	//sitemap结构体
	type data struct {
		Display string `xml:"display"`
	}
	type url struct {
		Loc        string  `xml:"loc"`        //网址
		Lastmod    string  `xml:"lastmod"`    //最后修改日期
		Changefreq string  `xml:"changefreq"` //更新频率
		Priority   float32 `xml:"priority"`   //比重 0.0 ~ 1.0
		Data       data    `xml:"data"`       //附加项
	}
	type urlset struct {
		Url []url `xml:"url"` //网站列表
	}

	//索引sitemap结构体
	type sitemap struct {
		Loc     string `xml:"loc"`     //网址
		Lastmod string `xml:"lastmod"` //最后修改日期
	}
	type sitemapindex struct {
		Sitemap []sitemap `xml:"sitemap"` //网站列表
	}

	//获取文章总数
	article := &models.Article{}
	articleCount := int(math.Ceil(float64(article.Count()) / float64(numPerXml))) //分页数

	//初始化索引sitemap
	var indexSitemap sitemapindex
	indexUrl := make([]sitemap, articleCount)

	for k := 0; k < articleCount; k++ { //遍历查询
		indexUrl[k].Loc = "http://www.wokugame.com/static/sitemap/sitemap_article" + strconv.Itoa((k + 1)) + ".xml"
		indexUrl[k].Lastmod = time.Now().Format("2006-01-02")
		//判断目录是否存在
		if _, err := os.Stat("static/sitemap/sitemap_article" + strconv.Itoa((k + 1)) + ".xml"); err == nil && k != articleCount-1 { //旧sitemap存在，并且不是最后一个
			continue
		}
		//查询这一段文章
		articles := article.FindSitemap(k*numPerXml, numPerXml)
		urls := make([]url, len(articles))
		//从最新的开始添加查询出的文章
		count := len(articles) - 1
		for j := count; j >= 0; j-- {
			urls[count-j].Loc = "http://www.wokugame.com/article/" + articles[j].Id.Hex() + ".html"
			urls[count-j].Priority = 0.6
			urls[count-j].Lastmod = articles[j].Id.Time().Format("2006-01-02")
			urls[count-j].Changefreq = "weekly"
		}
		var sitemap urlset
		sitemap.Url = urls
		//生成sitemap内容
		output, err := xml.Marshal(sitemap)
		if err != nil {
			continue
		}
		sitefile, _ := os.Create("static/sitemap/sitemap_article" + strconv.Itoa((k + 1)) + ".xml")
		sitefile.WriteString(xml.Header)
		sitefile.Write(output)
	}

	//生成索引sitemap
	indexSitemap.Sitemap = indexUrl
	output, err := xml.Marshal(indexSitemap)
	if err != nil {
		return
	}
	sitefile, _ := os.Create("static/sitemap/sitemap_index.xml")
	sitefile.WriteString(xml.Header)
	sitefile.Write(output)

	//向谷歌提交sitemap
	httplib.Get("http://www.google.com/webmasters/tools/ping?sitemap=http://www.wokugame.com/static/sitemap_index.xml").Response()
}

/* 判断是否为手机访问 */
func IsMobile(agent string) bool {
	//排除 Windows 桌面系统
	if !strings.Contains(agent, "Windows NT") || (strings.Contains(agent, "Windows NT") && strings.Contains(agent, "compatible; MSIE 9.0;")) {
		//排除 苹果桌面系统
		if !strings.Contains(agent, "Windows NT") && !strings.Contains(agent, "Macintosh") {
			for _, v := range mobileBrowserList {
				if strings.Contains(agent, v) {
					return true
				}
			}
		}
	}
	return false
}

/* ----------工具方法---------- */

/* 将[]string 转化为 []int */
func StringToIntArray(stringArray []string) []int {
	result := make([]int, len(stringArray))
	for k, _ := range stringArray {
		number, err := strconv.Atoi(stringArray[k])
		if err != nil { //转换错误
			return []int{}
		}
		result[k] = number
	}
	return result
}

/* 移除markdown标识 */
func RemoveMarkdown(content string) string {
	//html解码
	content = html.UnescapeString(content)
	//移除所有 * ` [ ] # - >
	content = strings.Replace(content, "*", "", -1)
	content = strings.Replace(content, "`", "", -1)
	content = strings.Replace(content, "[", "", -1)
	content = strings.Replace(content, "]", "", -1)
	content = strings.Replace(content, "#", "", -1)
	content = strings.Replace(content, "-", "", -1)
	content = strings.Replace(content, "> ", "", -1)
	//移除开头空格
	content = strings.Trim(content, "")
	return content
}

/*
 * 将新发布文章图片移动到指定文件夹
 * content *string 文章内容
 * prefix string 移动后前缀 eg: game/blog/123456000/123456000/123456000
 * @return int64 新增图片大小
 * @return string 更新地址后内容
 */
func HandleNewImage(content string, prefix string) (int64, string) {
	//匹配地址正则
	re, _ := regexp.Compile("\\([^()]+?(jpeg|jpg|png|gif|bmp)\\)")
	//正则将原文章内容所有图片地址都匹配出来
	images := re.FindAllString(content, -1)

	//取出路径
	for k, _ := range images {
		images[k] = strings.Replace(images[k], "http://img.wokugame.com/", "", -1)
		images[k] = strings.TrimLeft(images[k], "(")
		images[k] = strings.TrimRight(images[k], ")")
	}

	// 批量获取图片信息
	qiniu := &QiniuController{}
	infos := qiniu.GetInfos("woku", images)

	// 计算总大小
	var size int64

	for k, _ := range infos {
		// 跳过未找到文件
		if infos[k].Code != 200 {
			continue
		}

		// 大小累加
		size += infos[k].Data.Fsize
	}

	// 移动后文件路径
	keyDest := make([]string, len(images))

	// 为移动后文件路径赋值,同时替换内容所有图片地址
	for k, _ := range images {
		keyDest[k] = strings.Replace(images[k], "temp/", prefix, -1)

		content = strings.Replace(content, images[k], keyDest[k], -1)
	}

	qiniu.MoveFiles("woku", images, "woku", keyDest)

	return size, content
}

/*
 * 删除编辑后删除的图片（必须包含前缀路径保证是此文章的）
 * 移动编辑后新增的图片（并返回更新路径后的内容）
 * content *string 文章内容
 * prefix string 移动后前缀 eg: game/blog/123456000/123456000/123456000
 * @return int64 新增图片大小
 * @return string 更新地址后内容
 */
func HandleUpdateImage(oldContent string, newContent string, prefix string) (int64, string) {
	// 匹配地址正则
	re, _ := regexp.Compile("\\([^()]+?(jpeg|jpg|png|gif|bmp)\\)")
	// 正则将原、新文章内容所有图片地址都匹配出来
	oldImages := re.FindAllString(oldContent, -1)
	newImages := re.FindAllString(newContent, -1)

	// 操作数组
	oldSetType := mapset.NewSet()
	newSetType := mapset.NewSet()

	// 取出路径
	for k, _ := range oldImages {
		oldImages[k] = strings.Replace(oldImages[k], "http://img.wokugame.com/", "", 1)
		oldImages[k] = strings.TrimLeft(oldImages[k], "(")
		oldImages[k] = strings.TrimRight(oldImages[k], ")")
		oldSetType.Add(oldImages[k])
	}
	for k, _ := range newImages {
		newImages[k] = strings.Replace(newImages[k], "http://img.wokugame.com/", "", 1)
		newImages[k] = strings.TrimLeft(newImages[k], "(")
		newImages[k] = strings.TrimRight(newImages[k], ")")
		newSetType.Add(newImages[k])
	}

	// 求交集
	intersect := oldSetType.Intersect(newSetType)

	// 求删除的部分（原图片数组与交集求差集）
	deletes := oldSetType.Difference(intersect)
	deletesSlice := deletes.ToSlice()

	// 求新增部分（新图片数组与交集求差集）
	adds := newSetType.Difference(intersect)
	addsSlice := adds.ToSlice()

	// 将被删除和新增部分组成[]string
	deletesArray := make([]string, len(deletesSlice))
	for k, _ := range deletesSlice {
		deletesArray[k] = deletesSlice[k].(string)

		// 如果不包含前缀（说明不是此文章的图片），则不能删除
		if !strings.Contains(deletesArray[k], prefix) {
			deletesArray[k] = ""
		}
	}
	addsArray := make([]string, len(addsSlice))
	for k, _ := range addsSlice {
		addsArray[k] = addsSlice[k].(string)
	}

	// 将删除的图片删除
	qiniu := &QiniuController{}
	qiniu.DeleteFiles("woku", deletesArray)

	// 获取新增图片信息
	infos := qiniu.GetInfos("woku", addsArray)

	// 计算总大小
	var size int64

	for k, _ := range infos {
		// 跳过未找到文件
		if infos[k].Code != 200 {
			continue
		}

		// 大小累加
		size += infos[k].Data.Fsize
	}

	// 移动后文件路径
	keyDest := make([]string, len(addsArray))

	// 为移动后文件路径赋值,同时替换内容所有图片地址
	for k, _ := range addsArray {
		keyDest[k] = strings.Replace(addsArray[k], "temp/", prefix, 1)

		newContent = strings.Replace(newContent, addsArray[k], keyDest[k], 1)
	}

	qiniu.MoveFiles("woku", addsArray, "woku", keyDest)

	return size, newContent
}

/* 查找markdown语法中的图片
 * @params content 文章内容
 * @params number 查找数量
 */
func FindImages(content string, number int) []string {
	re, _ := regexp.Compile("\\([^()]+?(jpeg|jpg|png|gif|bmp)\\)")
	imageArray := re.FindAllString(content, number)
	//返回结果
	result := make([]string, number)
	for k, _ := range imageArray {
		//替换括号
		imageArray[k] = strings.Replace(imageArray[k], "(", "", 1)
		imageArray[k] = strings.Replace(imageArray[k], ")", "", 1)
		result[k] = imageArray[k]
	}
	return result
}

/* md5简化用法 */
func MD5(text string) string {
	m := md5.New()
	m.Write([]byte(text))
	return hex.EncodeToString(m.Sum(nil))
}

/* 查找是否有缓存
 * 先从内存中查找，如果没有，再从文件中查找
 */
func FindCache(key string) interface{} {
	if beego.RunMode != "prod" { //不是部署模式不会查找缓存
		return nil
	}
	var value interface{}
	//从内存中查找
	if value = Mc.Get(key); value != nil && value != "" {

	} else if value = Bm.Get(key); value != nil && value != "" {
		//重新设置到内存中
		Mc.Put(key, value, 300)
	} else {
		return nil
	}
	return value
}

/* 设置缓存
 * 将内容设置到内存缓存保存5分钟，同时设置到文件缓存保存1小时
 */
func SetCache(key string, value interface{}) {
	if beego.RunMode != "prod" { //不是部署模式不会设置缓存
		return
	}
	//设置内存缓存
	Mc.Put(key, value, 300)
	//设置文件缓存
	Bm.Put(key, value, 3600)
}

/* 某IP是否操作过于频繁 */
func IsBusy(name string, ip string, time int64) bool {
	if cac := Mc.Get(name + ip); cac != nil && cac != "" { //过于频繁
		return true
	} else {
		Mc.Put(name+ip, 1, time) //冷却时间
		return false
	}
}

/* 为某个用户增加一条消息
 * @params id string 用户id
 * @params _type string 类型
 * @params link string 链接地址
 * @params description string 描述
 * @params content string 全部内容
 */
func AddMessage(id string, _type string, category string, info string) {
	message := &models.Message{}
	message.Member = bson.ObjectIdHex(id)
	message.Type = _type
	message.Info = info
	message.Category = category
	message.Insert()
	//用户未读消息自增1，总消息数自增1
	member := models.Member{}
	member.FindOne(id)
	member.MessageNumber++
	member.MessageAll++
	//如果总消息数超过300，清理超过的100条
	if member.MessageAll >= 300 {
		message.ClearOverMessage(member.Id)
		member.MessageAll = 200
	}
	//socket推送一条消息
	if conn, ok := Clients[id]; ok {
		conn.WriteMessage(1, []byte(strconv.Itoa(member.MessageNumber)))
	}
	//存储信息
	member.Save()
}

/* 处理订单 */
func HandleOrder(buyerEmail string, tradeNo string, order *models.Order) bool {
	fmt.Println("处理账单开始")
	if ok := order.FindOne(); ok { //存在订单
		if order.Success == false { //如果订单处于未处理状态
			//查询该订单充值的用户
			member := &models.Member{}
			//根据订单类型处理信息
			switch order.Type {
			case "mobile": //手机支付 post访问其回调页面
				if order.Notify == "" {
					break
				}
				payString := strconv.FormatFloat(float64(order.Pay), 'f', 2, 32)
				u, _ := url.Parse(order.Notify + "?type=mobile_pay&statu=1&pay=" + payString + "&reqid=" + order.Reqid + "&game=" + order.Game + "&extend=" + order.Extend)
				q := u.Query()
				fmt.Println("encode后参数为：", q.Encode())
				sign := MD5(q.Encode() + member.Token)
				q.Add("sign", sign)
				u.RawQuery = q.Encode()
				req := httplib.Post(order.Notify)
				req.Param("statu", "1")
				req.Param("pay", payString)
				req.Param("reqid", order.Reqid)
				req.Param("game", order.Game)
				req.Param("extend", order.Extend)
				req.Param("sign", sign)
				req.String()
			}

			if time.Now().Sub(order.Time).Hours() <= 2 { //如果账单没有过期 2小时
				//重置订单为处理过状态，保存
				if ok = order.Update(bson.M{"$set": bson.M{"s": true, "aa": buyerEmail, "an": tradeNo, "t": time.Now()}}); ok { //保存成功
					//查询该订单充值的用户
					member := &models.Member{}
					if ok := member.FindOne(order.ToId.Hex()); ok { //有此用户
						//为用户充值
						member.Money += order.Pay
						//历史金额追加
						if member.MoneyHistory == "" {
							member.MoneyHistory = strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.FormatFloat(float64(member.Money), 'f', 2, 32)
						} else {
							member.MoneyHistory += ";" + strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.FormatFloat(float64(member.Money), 'f', 2, 32)
						}
						historyArray := strings.Split(member.MoneyHistory, ";")
						if len(historyArray) > 30 { //如果历史记录超过30条
							//只取前30条
							subArray := historyArray[len(historyArray)-30 : len(historyArray)]
							member.MoneyHistory = strings.Join(subArray, ";")
						}
						//保存用户信息
						member.Update(bson.M{"$set": bson.M{"mo": member.Money, "mh": member.MoneyHistory}})
						//发送邮件
						SendEmail([]string{member.Email}, "您的账号"+member.Nickname+"已成功充值"+strconv.FormatFloat(float64(order.Pay), 'f', 2, 32)+"元", `
							<p style="margin:0 0 35px;">
								尊敬的`+member.Nickname+`：
								<br>
								您的账户已成功充值 `+strconv.FormatFloat(float64(member.Money), 'f', 2, 32)+` 元。 请您登陆<a href='http://www.wokugame.com/user.html'>用户中心查看</a>
							</p>
							<p style="padding:20px 0;">
								祝您游戏愉快！
							</p>
							`)
						//根据订单类型处理信息
						switch order.Type {
						case "mobile": //手机支付 post访问其回调页面
							if order.Notify == "" {
								break
							}
							payString := strconv.FormatFloat(float64(order.Pay), 'f', 2, 32)
							u, _ := url.Parse(order.Notify + "?type=mobile_pay&statu=1&pay=" + payString + "&reqid=" + order.Reqid + "&game=" + order.Game + "&extend=" + order.Extend)
							q := u.Query()
							sign := MD5(q.Encode() + member.Token)
							q.Add("sign", sign)
							u.RawQuery = q.Encode()
							req := httplib.Post(order.Notify)
							req.Param("statu", "1")
							req.Param("pay", payString)
							req.Param("reqid", order.Reqid)
							req.Param("game", order.Game)
							req.Param("extend", order.Extend)
							req.Param("sign", sign)
							req.String()
						}
						return true
					} else {
						return false
					}
				} else { //并发问题，在这期间已经被其他回调方式修改，则不做处理
					return true
				}
			} else {
				return false
			}
		} else { //订单处理过，则不做处理
			return true
		}
	} else {
		return false
	}
	return false
}

/* 解析完整html代码中charset属性 */
func HtmlMetaCharset(html string) string {
	html = strings.ToLower(html)
	top := beego.Substr(html, 0, 2000) //取出足够多的头部
	start := strings.Index(top, "charset=")
	bs := []byte(top)[start+8:]
	out := string(bs)                                                                            //截取charset后面的字符串
	quoteStart := strings.Index(out, "\"")                                                       //找到第一个"
	ltStart := strings.Index(out, "<")                                                           //第一个<
	gtStart := strings.Index(out, ">")                                                           //第一个>
	eqStart := strings.Index(out, "=")                                                           //第一个=
	if quoteStart > -1 && quoteStart < ltStart && quoteStart < gtStart && quoteStart < eqStart { //"存在并且在< > =左边
		return beego.Substr(out, 0, quoteStart)
	}
	return ""
}

/* 根据hostUrl和imageUrl 返回图片的绝对路径 */
func ImageUrlAbs(hostUrl string, imageUrl string) string {
	hostArray := strings.Split(hostUrl, "/") //解析主机路径
	hostArray[len(hostArray)-1] = ""         //将主机路径的最后一位置空
	if !strings.Contains(imageUrl, "/") {    //如果没有路径，直接是地址
		returnUrl := ""
		for hostK, _ := range hostArray {
			returnUrl += hostArray[hostK]
			if hostK < len(hostArray)-1 {
				returnUrl += "/"
			}
		}
		returnUrl += "/" + imageUrl
		return returnUrl
	} else if !strings.Contains(imageUrl, "http://") { //如果图片没有http
		//如果以 / 开头，说明是绝对路径，直接加http://domain 否则是相对路径，解析
		imageArray := strings.Split(imageUrl, "/") //解析http://image.baidu.com/img/p/1.jpg
		for imK, _ := range imageArray {
			if imK == 0 && imageArray[imK] == "" { //如果第一个为空，说明是绝对路径
				//给图片地址添加完整的http://www.host.com
				return hostArray[0] + "/" + hostArray[1] + "/" + hostArray[2] + imageUrl
			}
			if imageArray[imK] != "http" && imageArray[imK] != "https" && imageArray[imK] != "" { //跳过http
				if imageArray[imK] == "." { //当级目录标识
					hostArray[len(hostArray)-1] = "" //将主机路径的最后一位置空
				} else if imageArray[imK] == ".." { //上级目录
					//将前一位置空，并将空主机路径最后一位删除
					hostArray[len(hostArray)-2] = ""
					hostArray = hostArray[:len(hostArray)-1]
				} else { //出现具体路径
					//主机完整路径+剩余具体路径
					imageUrl = ""
					for hostK, _ := range hostArray {
						imageUrl += hostArray[hostK]
						if hostK < len(hostArray)-1 {
							imageUrl += "/"
						}
					}
					copyImageArray := imageArray[imK:]
					for copyImK, _ := range copyImageArray {
						imageUrl += copyImageArray[copyImK]
						if copyImK < len(copyImageArray)-1 {
							imageUrl += "/"
						}
					}
					return imageUrl
				}
			}
		}
	}
	return imageUrl
}

/* 又拍云暂存图片 */
func UploadFlash(url string) (string, bool) {
	//远程下载该图片
	imageG := httplib.Get(url)
	imageR, err := imageG.Response()
	if err != nil {
		return "图片地址已失效", false
	}
	//如果是图片类型，获取后缀
	var suffix string
	imageCtype := imageR.Header.Get("Content-Type")
	imageTypeArray := strings.Split(imageCtype, "/")
	if imageTypeArray[0] == "image" { //是图片类型
		suffix = imageTypeArray[1]
	} else {
		return "这不是一个图片", false
	}
	//保存到临时文件夹
	imageName := bson.NewObjectId().Hex()
	filePath := "static/tmp/" + imageName + "." + suffix
	imageG.ToFile(filePath)
	//上传又拍云
	api := &ApiController{}
	ok, dir := api.UploadFile(filePath, "situation/flash", true)
	if !ok {
		return "服务器端上传失败", false
	}
	//列出目录，大于200个缓存图片则删除最后100个图片
	go func() {
		list, err := u.ReadDir("/situation/flash")
		if err == nil && len(list) > 200 {
			for k, l := 101, len(list); k < l; k++ {
				go u.DeleteFile("/situation/flash/" + list[k].Name)
			}
		}
	}()
	return dir, true
}

/* 结构体编码为字节流 */
func StructEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/* 字节流解码为结构体 */
func StructDecode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

/* 验证是否被恶意代理 */
func IsProxy(domain string) bool {
	if beego.RunMode == "prod" && domain != beego.AppConfig.String("webSite") && domain != beego.AppConfig.String("httpWebSite") {
		return true
	}
	return false
}

/* 判断是否在数组中 */
func inArray(key string, arr []string) bool {
	for k, _ := range arr {
		if arr[k] == key {
			return true
		}
	}
	return false
}

// 判断是否在int数组中
func inIntArray(key int, arr []int) bool {
	for k, _ := range arr {
		if arr[k] == key {
			return true
		}
	}
	return false
}
