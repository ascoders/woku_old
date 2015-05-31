package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"strings"
	"time"
	"woku/models"
)

type GameController struct {
	beego.Controller
}

var (
	noticeNumber int //公告每页文章数
	replyNumber  int //讨论每页回复数
	topNumber    int //最大置顶数量
)

func init() {
	noticeNumber = 20
	replyNumber = 20
	topNumber = 5
}

/* 获取游戏列表 */
func (this *GameController) GetGameList() {
	_type, _ := this.GetInt("type") //分类
	from, _ := this.GetInt("from")
	number, _ := this.GetInt("number")

	ok, data := func() (bool, interface{}) {
		if number > 100 {
			return false, "最多显示100项"
		}

		game := &models.Game{}
		games := game.Find(uint8(_type), from, number)
		count := game.FindCount(uint8(_type))

		return true, map[string]interface{}{
			"list":  games,
			"count": count,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 获取游戏基本信息 */
func (this *GameController) GetInfo() {
	game := &models.Game{}

	ok, data := func() (bool, interface{}) {
		if !game.FindPath(this.GetString("game")) {
			return false, nil
		}

		//查询该游戏的下属分类
		gameCategory := &models.GameCategory{}
		categorys := gameCategory.FindCategorys(game.Id)

		//遍历每个分类，查询前N个最新推送
		news := make([][]*models.Topic, len(categorys))
		topic := &models.Topic{}
		for k, _ := range categorys {
			// 推荐数量0则不查询
			if categorys[k].Recommend == 0 {
				news[k] = nil
				continue
			}

			news[k] = topic.FindNew(game.Id, categorys[k].Id, categorys[k].Recommend)
			for key, _ := range news[k] {
				news[k][key].Content = beego.Substr(news[k][key].Content, 0, 200)
			}
		}

		return true, map[string]interface{}{
			"Ok":        true,
			"Game":      game,      //游戏信息
			"News":      news,      //最新资讯（每个分类的）
			"Categorys": categorys, //分类信息
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 获取列表 */
func (this *GameController) GetList() {
	from, _ := this.GetInt("from")
	number, _ := this.GetInt("number")

	ok, data := func() (bool, interface{}) {
		if number > 100 {
			return false, "最多显示100项"
		}

		topic := &models.Topic{}

		//查询该分类列表
		articles := topic.Find(this.GetString("game"), this.GetString("category"), from, number)

		//找出前3个图片并取内容前200个字符
		images := make([][]string, len(articles))
		for k, _ := range articles {
			//查询每个文章前3个图片
			images[k] = FindImages(articles[k].Content, 3)
			//内容长度截取
			articles[k].Content = beego.Substr(articles[k].Content, 0, 200)
		}

		//查询该分类文章总数
		count := topic.FindCount(this.GetString("game"), this.GetString("category"))

		return true, map[string]interface{}{
			"articles": articles,
			"images":   images,
			"count":    count,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 获取文章容页 */
func (this *GameController) GetPage() {
	from, _ := this.GetInt("from")
	number, _ := this.GetInt("number")

	ok, data := func() (bool, interface{}) {
		if number > 100 {
			return false, "最多显示100个评论"
		}

		topic := &models.Topic{}
		if _ok := topic.FindOne(this.GetString("game"), this.GetString("id")); !_ok {
			return false, "文章不存在"
		}

		// 分词不需要提供
		topic.ContentSego = []string{}

		//增加浏览量
		topic.AddViews(this.GetString("id"))

		//查询回复
		reply := &models.Reply{}
		replys := reply.Find(topic.Id.Hex(), from, number)
		if from != 0 && replys == nil { //该页不存在
			return false, "该页不存在"
		}

		return true, map[string]interface{}{
			"topic":  topic,
			"replys": replys,
			"count":  topic.OutReply,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 查询嵌套评论 */
func (this *GameController) FindReply() {
	page, err := this.GetInt("page")
	if err != nil {
		this.StopRun()
	}
	//实例化评论
	reply := &models.Reply{}
	//查询该页最多5个评论
	replys := reply.FindReply(this.GetString("reply"), (page-1)*5, 5)

	this.Data["json"] = replys
	this.ServeJson()
}

/* 发帖 回帖 回复 编辑请求-post */
func (this *GameController) AddTopic() {
	ok, data := func() (bool, interface{}) {
		if len([]rune(this.GetString("title"))) > 60 {
			return false, "标题最多60个字符"
		}

		if len([]rune(this.GetString("content"))) > 10000 {
			return false, "内容最多10000个字符"
		}

		//未登录
		if this.GetSession("WOKUID") == nil {
			return false, "未登录"
		}

		//查询用户
		member := &models.Member{}
		if ok := member.FindOne(this.GetSession("WOKUID").(string)); !ok {
			return false, "用户不存在"
		}

		//获取游戏信息
		game := &models.Game{}
		if ok := game.FindPath(this.GetString("game")); !ok {
			return false, "该板块不存在"
		}

		//获取分类信息
		gameCategory := &models.GameCategory{}
		if ok := gameCategory.Find(this.GetString("game"), this.GetString("category")); !ok {
			return false, "该分类不存在"
		}

		// 文档模式允许内容为空
		if len([]rune(this.GetString("content"))) < 3 && gameCategory.Type != 1 {
			return false, "内容至少3个字符"
		}

		//3秒钟之内才能回帖
		if time.Now().Sub(member.LastTime).Seconds() < 1 {
			return false, "发帖间隔1秒"
		}
		member.UpdateFinish() //回帖时间重置

		switch this.GetString("type") {
		case "addTopic": //发帖
			if gameCategory.Type == 1 && strings.Trim(this.GetString("title"), " ") == "" {
				return false, "标题不能为空"
			}

			//判断是否有权限
			switch gameCategory.Add {
			case 0: //是否为管理员或版主
				if member.Id != game.Manager && !inArray(member.Id.Hex(), game.Managers) {
					return false, "管理员或版主才能发帖"
				}
				break
			case 1: //登陆即可，不做限制
				break
			}

			// 判断标签是否合法
			tags := this.GetStrings("tag")

			// 标签数量限制、长度限制、同文章不能重复
			if len(tags) >= 5 {
				return false, "最多5个标签"
			}

			for k, _ := range tags {
				if len([]rune(tags[k])) > 30 || len([]rune(tags[k])) < 1 {
					return false, "标签最大长度为30"
				}
			}

			copyTags := tags

			for k, _ := range tags {
				count := 0
				for _k, _ := range copyTags {
					if copyTags[_k] == tags[k] {
						count++

						if count >= 2 {
							return false, "标签不能重复"
						}
					}
				}
			}

			tag := &models.Tag{}
			// tag新增
			for k, _ := range tags {
				tag.Upsert(game.Id, tags[k])
			}

			// 实例化新增文章
			topic := &models.Topic{}
			topic.SetId() //设置id

			// 判断分类类型
			switch gameCategory.Type {
			case 0: // 论坛
			case 1: // 文档
				doc := &DocController{}

				isFolder, _ := this.GetBool("isFolder")

				// 新增文档
				tempOk := false
				tempData := ""
				if tempOk, tempData = doc.Add(this.GetString("title"), isFolder, this.GetString("docId"), this.GetString("docParent"), gameCategory.Id.Hex(), topic.Id.Hex()); !tempOk {
					return false, tempData
				}

				// 如果是目录，不进行后续插入文章操作
				if isFolder {
					return true, tempData
				}
			}

			var title string
			if this.GetString("title") != "" {
				title = this.GetString("title")
			} else {
				title = beego.Substr(this.GetString("content"), 0, 60)
			}

			// 文章赋值
			topic.Game = this.GetString("game")
			topic.Title = title
			topic.Category = gameCategory.Id
			topic.Content = this.GetString("content")
			topic.Ip = this.Ctx.Input.IP()
			topic.Author = member.Id
			topic.AuthorName = member.Nickname
			topic.AuthorImage = member.Image
			topic.LastReply = topic.Author
			topic.Tag = tags

			// 查找内容中所有图片信息，并移动到正确位置，同时替换图片路径
			var uploadSize int64
			uploadSize, topic.Content = HandleNewImage(topic.Content, "game/"+game.Id+"/"+topic.Id.Hex()+"/")

			// 生成内容分词表
			sego := &SegoController{}
			topic.ContentSego = sego.ToSlices(topic.Content, false)

			// 记录用户上传量
			member.UploadSize += uploadSize
			member.Save()

			//插入数据库
			objectId := topic.Insert()

			//TODO: 刷新缓存

			//游戏活跃度+1
			game.Hot++

			//保存板块修改
			game.SaveAll()

			return true, objectId.Hex()
		case "reply": //回复
			//判断是否有权限
			switch gameCategory.Add {
			case 0: //是否为管理员或版主
				if member.Id != game.Manager && !inArray(member.Id.Hex(), game.Managers) {
					return false, "管理员或版主才能发帖"
				}
				break
			case 1: //登陆即可，不做限制
				break
			}

			if len([]rune(this.GetString("content"))) > 5000 {
				return false, "回帖最多5000个字符"
			}

			//查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("topic")); !ok { //不存在此话题
				return false, "回复的帖子不存在"
			}

			//回复数自增1,更新最后一个回复者的id
			topic.Reply++
			topic.OutReply++
			topic.LastReply = member.Id
			topic.LastReplyName = member.Nickname
			topic.LastTime = time.Now()
			topic.Save()

			//实例化回复
			reply := &models.Reply{}
			reply.Game = this.GetString("game")
			reply.Topic = this.GetString("topic")
			reply.Content = this.GetString("content")
			reply.Author = member.Id
			reply.AuthorName = member.Nickname
			reply.AuthorImage = member.Image
			reply.Ip = this.Ctx.Input.IP()
			reply.SetId()

			// 查找内容中所有图片信息，并移动到正确位置，同时替换图片路径
			var uploadSize int64
			uploadSize, reply.Content = HandleNewImage(reply.Content, "game/"+game.Id+"/"+topic.Id.Hex()+"/"+reply.Id.Hex()+"/")

			// 记录用户上传量
			member.UploadSize += uploadSize
			member.Save()

			//插入回复
			reply.Insert()

			//推送消息，如果操作用户不是回复作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/g/"+this.GetString("game")+"/"+this.GetString("topic")+".html?p="+this.GetString("page")+"#"+reply.Id.Hex(), "您的一个帖子有了新回复", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;回复了您的帖子：<span class='text-muted'>&nbsp"+reply.Content+"&nbsp;</span>", "")
			}

			//游戏活跃度+1
			game.Hot++
			game.SaveAll()

			return true, reply.Id.Hex()
		case "rReply": //回复评论
			//判断是否有权限
			switch gameCategory.Add {
			case 0: //是否为管理员或版主
				if member.Id != game.Manager && !inArray(member.Id.Hex(), game.Managers) {
					return false, "管理员或版主才能发帖"
				}
				break
			case 1: //登陆即可，不做限制
				break
			}

			if len([]rune(this.GetString("content"))) > 500 {
				return false, "追加评论最多500个字符"
			}

			//查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("topic")); !ok { //不存在此话题
				return false, "帖子不存在"
			}

			//查找回复
			reply := &models.Reply{}
			if ok := reply.FindById(this.GetString("reply")); ok == false { //没有查找到回复则退出
				return false, "回复不存在"
			}

			//实例化嵌套评论
			newReply := &models.Reply{}
			newReply.Game = this.GetString("game")
			newReply.Topic = this.GetString("topic")
			newReply.Reply = reply.Id.Hex()
			newReply.Content = this.GetString("content")
			newReply.Author = member.Id
			newReply.AuthorName = member.Nickname
			newReply.AuthorImage = member.Image
			newReply.Ip = this.Ctx.Input.IP()
			newReply.SetId()

			//插入嵌套评论
			newReply.Insert()

			//刷新前五个嵌套评论
			reply.FreshCache()

			//回复的评论数量+1
			reply.ReplyNumber++
			reply.Save()

			//帖子回复数自增1,更新最后一个回复者的id
			topic.Reply++
			topic.LastReply = member.Id
			topic.LastReplyName = member.Nickname
			topic.LastTime = time.Now()

			//保存
			topic.Save()

			//TODO:刷新缓存

			//推送消息，如果操作用户不是回复作者的话，通知他
			if reply.Author != member.Id {
				//AddMessage(reply.Author.Hex(), "game", "/g/"+this.GetString("category")+"/"+this.GetString("topic")+"#"+newReply.Id.Hex(), "您的一个回复有了新评论", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;评论了您的回复：<span class='text-muted'>&nbsp"+newReply.Content+"&nbsp;</span>", "")
			}

			return true, newReply.Id.Hex()
		case "edit": //编辑
			//查询该话题
			topic := &models.Topic{}
			topic.FindById(this.GetString("topic"))

			if topic == nil { //话题不存在
				return false, "帖子不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			// 删除编辑后删除的图片（必须包含前缀路径保证是此文章的）
			// 移动编辑后新增的图片（并返回更新路径后的内容）
			var uploadSize int64
			uploadSize, topic.Content = HandleUpdateImage(topic.Content, this.GetString("content"), "game/"+game.Id+"/"+topic.Id.Hex()+"/")

			// 生成内容分词表
			sego := &SegoController{}
			topic.ContentSego = sego.ToSlices(topic.Content, false)

			// 记录用户上传量
			member.UploadSize += uploadSize
			member.Save()

			//保存
			topic.Save()

			//TODO:刷新缓存

			//推送消息，如果操作用户不是回复作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/game/"+this.GetString("game")+"/"+this.GetString("topic")+".html", "您有一个帖子被管理员编辑", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;编辑了您的帖子：<span class='text-muted'>&nbsp"+topic.Title+"&nbsp;</span>", "")
			}
			break
		}

		return true, ""
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 将帖子移动到其他分类下 */
func (this *GameController) ChangeCategory() {
	ok, data := func() (bool, interface{}) {
		//未登录
		if this.GetSession("WOKUID") == nil {
			return false, "未登录"
		}

		//查询用户
		member := &models.Member{}
		if ok := member.FindOne(this.GetSession("WOKUID").(string)); !ok {
			return false, "用户不存在"
		}

		//查询该话题
		topic := &models.Topic{}
		if _ok := topic.FindById(this.GetString("topic")); !_ok {
			return false, "帖子不存在"
		}

		//获取获得游戏信息
		game := &models.Game{}
		if _ok := game.FindPath(topic.Game); !_ok {
			return false, "该板块不存在"
		}

		//获取分类信息
		gameCategory := &models.GameCategory{}
		if _ok := gameCategory.Find(topic.Game, this.GetString("category")); !_ok {
			return false, "该分类不存在"
		}

		if gameCategory.Game != topic.Game {
			return false, "该分类不在此板块中"
		}

		if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
			return false, "无权限"
		}

		// 修改分类
		topic.Category = gameCategory.Id

		// 保存
		topic.Save()

		//TODO:刷新缓存

		return true, ""
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 上传图片 */
func (this *GameController) UploadHandle() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		game := &models.Game{}
		if ok := game.FindPath(this.GetString("game")); !ok {
			return false, "板块不存在"
		}

		// 获取图片信息
		qiniu := &QiniuController{}
		__ok, info := qiniu.GetInfo("woku", this.GetString("name"))
		if !__ok {
			return false, "文件不存在"
		}

		var remotePath string

		//根据类别检查权限
		switch this.GetString("type") {
		case "gameImage": //应用图标
			if member.Id != game.Manager {
				return false, "没有权限"
			}

			//如果已有图片，删除旧图片
			if game.GameImage != "" {
				qiniu.DeleteFile("woku", game.GameImage)
			} else { //活跃度加2
				game.Hot += 2
			}

			game.GameImage = "game/" + game.Id + "/admin/image" + this.GetString("ext")
			//保存游戏信息
			game.SaveAll()

			// 移动文件
			qiniu.MoveFile("woku", this.GetString("name"), "woku", game.GameImage)

			remotePath = game.GameImage
		case "gameIcon": //应用ico
			if member.Id != game.Manager {
				return false, "没有权限"
			}

			//如果已有图片，删除旧图片
			if game.Icon != "" {
				qiniu.DeleteFile("woku", game.Icon)
			} else { //活跃度加2
				game.Hot += 2
			}

			game.Icon = "game/" + game.Id + "/admin/icon" + this.GetString("ext")
			//保存游戏信息
			game.SaveAll()

			// 移动文件
			qiniu.MoveFile("woku", this.GetString("name"), "woku", game.Icon)

			remotePath = game.Icon
		case "gameScreenShot": //截图展示
			if member.Id != game.Manager {
				return false, "没有权限"
			}

			//获取插入位置
			position, _ := this.GetInt("position")
			if position < 0 || position >= 6 { //不在范围内
				return false, "截图序号错误，不在范围内"
			}

			// 设置移动后路径
			remotePath = "game/" + game.Id + "/admin/screenShot" + this.GetString("position") + this.GetString("ext")

			//如果超过长度，增加
			if int(position) >= len(game.Image) {
				game.Image = append(game.Image, remotePath)
				game.Hot += 2
			} else {
				//如果已有图片，删除旧图片
				if game.Image[position] != "" {
					qiniu.DeleteFile("woku", game.Image[position])
				} else { //没有图片
					game.Hot += 2
				}
				//保存新图片信息
				game.Image[position] = remotePath
			}

			//移动图片
			qiniu.MoveFile("woku", this.GetString("name"), "woku", remotePath)

			//保存游戏信息
			game.SaveAll()
		default: //请求类型不在范围内
			return false, "无效的请求"
		}

		//清空game缓存
		models.DeleteCache("game-findpath-" + game.Id)

		//如果用户最后上传日期不等于今天，重新赋值
		if member.UploadDate.Year() < time.Now().Year() || member.UploadDate.YearDay() < time.Now().YearDay() { //天数小于今天
			member.UploadDate = time.Now()
			member.UploadSize = info.Fsize
		} else { //是今天则累加
			member.UploadSize += info.Fsize
		}
		member.Save()

		return true, remotePath
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 管理-基本信息保存 */
func (this *GameController) BaseSave() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if _ok := member.FindOne(session.(string)); !_ok {
			return false, "用户不存在"
		}

		game := &models.Game{}
		if _ok := game.FindPath(this.GetString("game")); !_ok {
			return false, "板块不存在"
		}

		if member.Id != game.Manager {
			return false, "没有权限"
		}

		size, _ := this.GetFloat("size")
		version, _ := this.GetFloat("version")
		game.Size = float32(size)
		game.Version = float32(version)
		game.Need = this.GetString("need")
		game.Description = this.GetString("description")
		game.Download = this.GetString("download")
		game.SaveAll() //保存数据

		//刷新缓存
		models.DeleteCache("game-FindPath-" + game.Id)
		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 管理 分类管理 修改推荐优先级 */
func (this *GameController) ChangeRecommendPri() {
	ok, data := func() (bool, interface{}) {
		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		member := &models.Member{}
		if _ok := member.FindOne(session.(string)); !_ok {
			return false, "用户不存在"
		}

		game := &models.Game{}
		if _ok := game.FindPath(this.GetString("game")); !_ok {
			return false, "板块不存在"
		}

		// 修改分类优先级
		category := &models.GameCategory{}
		value, _ := this.GetInt("value")
		if _ok, _data := category.ChangeRecommendPri(this.GetString("game"), this.GetString("category"), value); !_ok {
			return false, _data
		}

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 申请创建游戏 */
func (this *GameController) CreateGame() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		if time.Now().Sub(member.LastTime).Seconds() < 2.9 { //3秒内不能连续操作
			return false, "请继续等待"
		}

		if len([]rune(this.GetString("name"))) < 2 || len([]rune(this.GetString("name"))) > 20 { //名称不能超过10个字符
			return false, "名称长度2-20"
		}

		if this.GetString("name") == "" {
			return false, "名称不能为空"
		}

		if len([]rune(this.GetString("path"))) < 3 || len([]rune(this.GetString("path"))) > 20 { //路径不能超过10个字符
			return false, "域名长度3-20"
		}

		if this.GetString("path") == "" {
			return false, "域名不能为空"
		}

		if this.GetString("path") == "post" || this.GetString("path") == "api" {
			return false, "该域名不能使用"
		}

		//正则验证
		re := regexp.MustCompile("[`~!！@#$%^&*()_+<>?:\"”{},.，。/;；‘'[\\]]")
		result := re.FindString(this.GetString("name"))
		if result != "" {
			return false, "名称不能有特殊字符"
		}

		re = regexp.MustCompile("[a-z]+")
		result = re.FindString(this.GetString("path"))
		if this.GetString("path") != result {
			return false, "域名只能输入字母"
		}

		game := &models.Game{}
		//查询名称是否重复
		if game.FindRepeatName(this.GetString("name")) {
			return false, "名称已存在"
		}

		//查询域名是否重复
		if game.FindRepeat(this.GetString("path")) {
			return false, "域名已存在"
		}

		//检测type只能是0~4
		_type, _ := this.GetInt("type")
		if _type != 0 && _type != 1 && _type != 2 && _type != 3 && _type != 4 {
			return false, "分类不在范围内"
		}

		//一个用户最多建立20个游戏
		if member.GameNumber >= 20 {
			return false, "抱歉，您最多创建20个游戏"
		}

		member.GameNumber++
		member.Save()

		//注册新游戏
		game.Name = this.GetString("name")
		game.Id = this.GetString("path")
		game.Type = uint8(_type)
		game.Manager = member.Id
		game.Categorys = 3
		game.Insert()

		//给游戏注册默认分类
		defaultCategory := map[string]string{
			"topic": "公告",
			"plan":  "攻略",
			"bbs":   "论坛",
		}

		// 存储层级
		layer := 0

		for k, _ := range defaultCategory {
			gameCategory := &models.GameCategory{}
			gameCategory.Game = game.Id
			gameCategory.Category = k
			gameCategory.CategoryName = defaultCategory[k]
			gameCategory.RecommendPri = layer
			gameCategory.Recommend = 5
			gameCategory.Add = 1
			gameCategory.Reply = 1
			gameCategory.Insert()

			layer++
		}

		// 删除此游戏查询缓存
		models.DeleteCache("game-findpath-" + this.GetString("path"))

		// 删除查询游戏列表缓存
		models.DeleteCaches("game-find-" + this.GetString("type"))
		models.DeleteCaches("game-findcount-" + this.GetString("type"))

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 新增分类 */
func (this *GameController) AddCategory() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		if len([]rune(this.GetString("name"))) < 1 || len([]rune(this.GetString("name"))) > 10 {
			return false, "名称长度1-10"
		}

		if len([]rune(this.GetString("path"))) < 1 || len([]rune(this.GetString("path"))) > 20 {
			return false, "路径长度1-20"
		}

		re := regexp.MustCompile("[a-z]+")
		result := re.FindString(this.GetString("path"))
		if this.GetString("path") != result {
			return false, "路径只能输入字母"
		}

		//路径名不能等于 admin
		if this.GetString("path") == "admin" {
			return false, "路径不能为admin"
		}

		//路径名不能等于 tag
		if this.GetString("path") == "tag" {
			return false, "路径不能为tag"
		}

		add, _ := this.GetInt("add")
		if add != 0 && add != 1 {
			return false, "发帖限制不在范围内"
		}

		number, _ := this.GetInt("number") //推荐数量
		if number < 0 || number > 20 {
			return false, "推荐数量0-20"
		}

		reply, _ := this.GetInt("reply")
		if reply != 0 && reply != 1 {
			return false, "回帖限制不在范围内"
		}

		pri, _ := this.GetInt("pri")
		if pri < 0 || pri > 20 {
			return false, "优先级0-20"
		}

		_type, _ := this.GetInt("_type")
		if _type < 0 || _type > 1 {
			return false, "类型不在范围内"
		}

		// 查找游戏
		game := &models.Game{}
		if _ok := game.FindPath(this.GetString("game")); !_ok {
			return false, "板块不存在"
		}

		// 检查权限
		if member.Id != game.Manager {
			return false, "需要管理员权限"
		}

		if game.Categorys >= 20 {
			return false, "超过20个模块"
		}

		// 路径是否重复
		gameCategory := &models.GameCategory{}
		if _ok := gameCategory.Find(this.GetString("game"), this.GetString("path")); _ok {
			return false, "路径重复"
		}

		// 游戏分类+1并保存
		game.Categorys++
		game.SaveAll()

		// 新增分类
		gameCategory.Game = game.Id
		gameCategory.Category = this.GetString("path")
		gameCategory.CategoryName = this.GetString("name")
		gameCategory.RecommendPri = pri
		gameCategory.Recommend = number
		gameCategory.Add = add
		gameCategory.Reply = reply
		gameCategory.Type = _type
		gameCategory.Insert()

		// 删除获取游戏的缓存(刷新当前分类数量)
		models.DeleteCache("game-findpath-" + game.Id)

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 更新分类信息 */
func (this *GameController) UpdateCategory() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		if len([]rune(this.GetString("name"))) < 1 || len([]rune(this.GetString("name"))) > 10 {
			return false, "名称长度1-10"
		}

		if len([]rune(this.GetString("path"))) < 1 || len([]rune(this.GetString("path"))) > 20 {
			return false, "路径长度1-20"
		}

		re := regexp.MustCompile("[a-z]+")
		result := re.FindString(this.GetString("path"))
		if this.GetString("path") != result {
			return false, "路径只能输入字母"
		}

		//路径名不能等于 admin
		if this.GetString("path") == "admin" {
			return false, "路径不能为admin"
		}

		add, _ := this.GetInt("add")
		if add != 0 && add != 1 {
			return false, "发帖限制不在范围内"
		}

		number, _ := this.GetInt("number") //推荐数量
		if number < 0 || number > 20 {
			return false, "推荐数量0-20"
		}

		reply, _ := this.GetInt("reply")
		if reply != 0 && reply != 1 {
			return false, "回帖限制不在范围内"
		}

		// 查找游戏
		game := &models.Game{}
		if _ok := game.FindPath(this.GetString("game")); !_ok {
			return false, "板块不存在"
		}

		// 检查权限
		if member.Id != game.Manager {
			return false, "需要管理员权限"
		}

		// 路径是否重复
		gameCategory := &models.GameCategory{}
		if _ok := gameCategory.Find(game.Id, this.GetString("path")); _ok {
			// 路径重复,并且不是当前修改的路径
			if gameCategory.Id.Hex() != this.GetString("id") {
				return false, "路径重复"
			}
		}

		_type, _ := this.GetInt("_type")
		if _type < 0 || _type > 1 {
			return false, "类型不在范围内"
		}

		// 查询分类信息
		if _ok := gameCategory.Find(game.Id, this.GetString("id")); !_ok {
			return false, "分类不存在"
		}

		// 切换分类需要清空其下文章
		if gameCategory.Type != _type {
			topic := &models.Topic{}
			if _ok := topic.HasTopic(game.Id, gameCategory.Id.Hex()); _ok {
				return false, "修改分类前请清空文章"
			}
		}

		// 根据id更新分类
		if _ok, _data := gameCategory.Update(this.GetString("id"), this.GetString("path"), this.GetString("name"), number, add, reply, _type); !_ok {
			return false, _data
		}

		// 删除获取游戏的缓存(刷新game.分类数量和分类信息)
		models.DeleteCache("game-FindPath-" + game.Id)

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 删除分类信息 */
func (this *GameController) DeleteCategory() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		// 查找游戏
		game := &models.Game{}
		if _ok := game.FindPath(this.GetString("game")); !_ok {
			return false, "板块不存在"
		}

		// 分类不能少于1
		if game.Categorys <= 1 {
			return false, "至少有1个分类"
		}

		// 检查权限
		if member.Id != game.Manager {
			return false, "需要管理员权限"
		}

		// 如果该分类下有文章，则不能删除
		topic := &models.Topic{}
		if _ok := topic.HasTopic(game.Id, this.GetString("id")); _ok {
			return false, "删除分类前请清空帖子"
		}

		// 根据id删除分类
		gameCategory := &models.GameCategory{}
		if _ok, _data := gameCategory.Delete(game.Id, this.GetString("id")); !_ok {
			return false, _data
		}

		// 游戏分类-1并保存
		game.Categorys--
		game.SaveAll()

		// 删除获取游戏的缓存(刷新game.分类数量和分类信息)
		models.DeleteCache("game-FindPath-" + game.Id)

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 对帖子的各项操作POST */
func (this *GameController) Operate() {
	ok, data := func() (bool, interface{}) {
		//查询用户
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if ok := member.FindOne(session.(string)); !ok {
			return false, "用户不存在"
		}

		_ok := false
		models.AutoCache(member.Id.Hex()+"operate", "1", 1, func() {
			_ok = true
		})
		if !_ok {
			return false, "操作过于频繁，请等待1秒"
		}

		switch this.GetString("type") {
		case "top": //置顶操作
			// 查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("id")); !ok { //不存在此话题
				return false, "回复的帖子不存在"
			}

			// 获取分类信息
			gameCategory := &models.GameCategory{}
			if ok := gameCategory.Find(topic.Game, topic.Category.Hex()); !ok {
				return false, "该分类不存在"
			}

			// 获取游戏信息
			game := &models.Game{}
			if _ok := game.FindPath(topic.Game); !_ok {
				return false, "板块不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			_ok := false
			models.AutoCache(member.Id.Hex()+topic.Id.Hex()+"topCancle", "1", 30, func() {
				_ok = true
			})
			if !_ok {
				return false, "您刚刚对这个帖子取消置顶过，请等待30秒再操作"
			}

			_ok, number := topic.AddTop(topic.Id.Hex(), topic.Game, topic.Category, topNumber)

			if !_ok {
				return false, "最多置顶5个"
			}

			//TODO:刷新缓存

			//推送消息，如果操作用户不是帖子作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/game/"+game.Id+"/"+topic.Id.Hex()+".html", "您在&nbsp;"+game.Name+"&nbsp;的一篇帖子被置顶", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;对您的帖子&nbsp;<span class='text-info'>"+topic.Title+"</span>&nbsp;执行了置顶操作，恭喜你啦，快去看看吧。", "")
			}

			return true, number
		case "topCancle": //取消置顶操作
			// 查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("id")); !ok { //不存在此话题
				return false, "回复的帖子不存在"
			}

			// 获取分类信息
			gameCategory := &models.GameCategory{}
			if ok := gameCategory.Find(topic.Game, topic.Category.Hex()); !ok {
				return false, "该分类不存在"
			}

			// 获取游戏信息
			game := &models.Game{}
			if _ok := game.FindPath(topic.Game); !_ok {
				return false, "板块不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			topic.CancleTop(topic.Id.Hex())

			//TODO:刷新缓存

			//推送消息，如果操作用户不是帖子作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/game/"+game.Id+"/"+topic.Id.Hex()+".html", "您在&nbsp;"+game.Name+"&nbsp;的一篇帖子被取消置顶", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;将您的帖子&nbsp;<span class='text-info'>"+topic.Title+"</span>&nbsp;取消置顶了哦。", "")
			}
			return true, 0
		case "good": //加精华操作
			// 查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("id")); !ok { //不存在此话题
				return false, "回复的帖子不存在"
			}

			// 获取分类信息
			gameCategory := &models.GameCategory{}
			if ok := gameCategory.Find(topic.Game, topic.Category.Hex()); !ok {
				return false, "该分类不存在"
			}

			// 获取游戏信息
			game := &models.Game{}
			if _ok := game.FindPath(topic.Game); !_ok {
				return false, "板块不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			_ok := false
			models.AutoCache(member.Id.Hex()+topic.Id.Hex()+"goodCancle", "1", 30, func() {
				_ok = true
			})
			if !_ok {
				return false, "您刚刚对这个帖子取消加精过，请等待30秒再操作"
			}

			topic.AddGood(topic.Id.Hex())

			//TODO:刷新缓存

			//推送消息，如果操作用户不是帖子作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/game/"+game.Id+"/"+topic.Id.Hex()+".html", "您在&nbsp;"+game.Name+"&nbsp;的一篇帖子被加精", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;对您的帖子&nbsp;<span class='text-info'>"+topic.Title+"</span>&nbsp;执行了加精操作，恭喜你啦，快去看看吧。", "")
			}
		case "goodCancle": //取消精华
			// 查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("id")); !ok { //不存在此话题
				return false, "回复的帖子不存在"
			}

			// 获取分类信息
			gameCategory := &models.GameCategory{}
			if ok := gameCategory.Find(topic.Game, topic.Category.Hex()); !ok {
				return false, "该分类不存在"
			}

			// 获取游戏信息
			game := &models.Game{}
			if _ok := game.FindPath(topic.Game); !_ok {
				return false, "板块不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			topic.CancleGood(topic.Id.Hex())

			//TODO:刷新缓存

			//推送消息，如果操作用户不是帖子作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "/game/"+game.Id+"/"+topic.Id.Hex()+".html", "您在&nbsp;"+game.Name+"&nbsp;的一篇帖子被取消加精", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;对您的帖子&nbsp;<span class='text-info'>"+topic.Title+"</span>&nbsp;取消了加精。", "")
			}
		case "delete": //删除帖子
			// 查找话题
			topic := &models.Topic{}
			if ok := topic.FindById(this.GetString("id")); !ok { //不存在此话题
				return false, "删除的帖子不存在"
			}

			// 获取分类信息
			gameCategory := &models.GameCategory{}
			if ok := gameCategory.Find(topic.Game, topic.Category.Hex()); !ok {
				return false, "该分类不存在"
			}

			// 获取游戏信息
			game := &models.Game{}
			if _ok := game.FindPath(topic.Game); !_ok {
				return false, "板块不存在"
			}

			if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) { //既不是管理组也不是作者
				return false, "无权限"
			}

			// 如果分类是文档类型，先删除文档下文件
			if gameCategory.Type == 1 {
				docContro := &DocController{}
				index, _ := this.GetInt("docIndex")
				if _ok, _data := docContro.DeleteFile(this.GetString("docParent"), index, topic.Id.Hex()); !_ok {
					return false, _data
				}
			}

			//删除帖子下全部评论和嵌套评论
			reply := &models.Reply{}
			reply.DeleteTopicReply(topic.Id.Hex())

			//删除图片
			qiniu := &QiniuController{}
			go qiniu.DeleteAll("woku", "game/"+game.Id+"/"+topic.Id.Hex()+"/")

			//删除帖子中的tag
			tag := &models.Tag{}
			for k, _ := range topic.Tag {
				tag.CountReduce(game.Id, topic.Tag[k])
			}

			//删除帖子
			topic.Delete()

			//TODO:刷新缓存

			//推送消息，如果操作用户不是帖子作者的话，通知他
			if topic.Author != member.Id {
				//AddMessage(topic.Author.Hex(), "game", "", "您在&nbsp;"+game.Name+"&nbsp;的一篇帖子被管理员删除", "<img userImage='"+member.Image+"'>&nbsp;<span class='text-info'>"+member.Nickname+"</span>&nbsp;将您的文章&nbsp;<span class='text-info'>"+topic.Title+"</span>&nbsp;删除，很遗憾哦。", "")
			}
		case "deleteReply": //删除回复/嵌套评论
			// 回帖删除权限有：管理组 > 帖子作者 > 回复者
			// 嵌套评论删除权限有： 管理组 > 帖子作者 > 回复者 > 回复评论者

			//查询该回复/嵌套评论
			reply := &models.Reply{}
			if !reply.FindById(this.GetString("reply")) {
				return false, "删除的回复不存在"
			}

			//查看该回复所属帖子
			topic := &models.Topic{}
			if !topic.FindById(reply.Topic) { //帖子不存在
				return false, "帖子不存在"
			}

			//查询该回复的所属板块
			game := &models.Game{}
			if ok := game.FindPath(reply.Game); !ok {
				return false, "该板块不存在"
			}

			//如果是嵌套评论，查询父级回复
			pReply := &models.Reply{}
			if reply.Reply != "" {
				if !pReply.FindById(reply.Reply) {
					return false, "父级回复不存在"
				}
			}

			// 权限审核
			// 直接作者放行
			// 管理员无条件放行
			// 版主（辅助管理组）无条件放行
			// 帖子作者无条件放行
			// 如果是嵌套回复，那么父级回帖的作者也会放行
			can := false
			can = can || member.Id == reply.Author
			can = can || member.Id == game.Manager
			can = can || inArray(member.Id.Hex(), game.Managers)
			can = can || member.Id == topic.Author
			can = can || member.Id == pReply.Author
			if !can {
				return false, "无权限"
			}

			//话题评论数-1
			topic.Reply--

			if reply.Reply == "" { //如果是最外层回复
				//最外层回复数-1
				topic.OutReply--

				//删除回复下全部评论
				change := reply.DeleteReplyReply(reply.Id.Hex())

				//删除该回复下所有图片
				qiniu := &QiniuController{}
				go qiniu.DeleteAll("woku", "game/"+game.Id+"/"+topic.Category.Hex()+"/"+topic.Id.Hex()+"/"+reply.Id.Hex()+"/")

				//话题评论数再减去删除的嵌套评论数
				topic.Reply -= change.Removed
			}

			//删除该回复/嵌套评论
			reply.Delete()

			//帖子保存
			topic.Save()

			if reply.Reply != "" { //如果是嵌套评论
				//父级回复评论数量-1，并刷新其评论缓存
				pReply.ReplyNumber--
				pReply.FreshCache()
				pReply.Save()
			}

			//todo:刷新缓存

			//推送消息，如果操作用户不是评论作者的话，通知他
			/*
				if member.Id != reply.Author {
					AddMessage(rReply.Author.Hex(), "game", url+"#"+reply.Id.Hex(), "您有一条评论被删除", `
						<img userImage='"+member.Image+"'>&nbsp;
						<span class='text-info'>"+member.Nickname+"</span>&nbsp;
						删除了您的评论：<span class='text-muted'>&nbsp
						"+rReply.Content+"&nbsp;</span>`, "")
				}
			*/
		}
		return true, ""
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

/* 判断用户是否是游戏管理员 */
func (this *GameController) isManager(contro *beego.Controller) (bool, *models.Game, *models.Member) {
	//获取游戏信息
	game := &models.Game{}
	ok := game.FindPath(contro.GetString("category"))
	if !ok {
		return false, nil, nil
	}
	//获取用户信息
	member := &models.Member{}
	if session := contro.GetSession("WOKUID"); session != nil {
		if ok := member.FindOne(session.(string)); ok {
			//如果不是这个游戏的管理员则无权限
			if member.Id != game.Manager {
				return false, nil, nil
			} else {
				return true, game, member
			}
		} else {
			return false, nil, nil
		}
	} else {
		return false, nil, nil
	}
}

/* 根据话题id判断是否为其管理员 */
func (this *GameController) isManagerById(contro *beego.Controller) (bool, *models.Topic, *models.Game, *models.Member) {
	//实例化话题
	topic := &models.Topic{}
	ok := topic.FindById(contro.GetString("id"))
	if !ok { //不存在此话题
		return false, nil, nil, nil
	}
	//获取游戏信息
	game := &models.Game{}
	ok = game.FindPath(topic.Game)
	if ok { //存在此游戏
		//获取用户信息
		member := &models.Member{}
		if session := contro.GetSession("WOKUID"); session != nil {
			if ok := member.FindOne(session.(string)); ok {
				//如果是游戏管理员
				if member.Id == game.Manager {
					return true, topic, game, member
				} else {
					return false, topic, game, member
				}
			}
		}
	}
	return false, nil, nil, nil
}
