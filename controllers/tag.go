package controllers

import (
	"woku/models"

	"github.com/astaxie/beego"
)

type TagController struct {
	beego.Controller
}

// 绑定标签
// topic string 所属文章
// name string 标签名
func (this *TagController) Bind() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if _ok := member.FindOne(session.(string)); !_ok {
			return false, "用户不存在"
		}

		topic := &models.Topic{}
		if !topic.FindById(this.GetString("topic")) {
			return false, "文章不存在"
		}

		game := &models.Game{}
		if _ok := game.FindPath(topic.Game); !_ok {
			return false, "板块不存在"
		}

		// 需要权限为作者或管理组
		if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) {
			return false, "没有权限"
		}

		// 标签数量限制、长度限制、同文章不能重复
		if len(topic.Tag) >= 5 {
			return false, "最多5个标签"
		}

		if len([]rune(this.GetString("name"))) > 30 || len([]rune(this.GetString("name"))) < 1 {
			return false, "标签最大长度为30"
		}

		tag := &models.Tag{}
		for k, _ := range topic.Tag {
			if this.GetString("name") == topic.Tag[k] {
				return false, "标签不能重复"
			}
		}

		// tag新增
		tag.Upsert(game.Id, this.GetString("name"))

		// 文章新增tag
		topic.AddTag(this.GetString("name"))

		return true, ""
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 解绑标签
// topic string 所属文章
// name string 标签名
func (this *TagController) UnBind() {
	ok, data := func() (bool, interface{}) {
		member := &models.Member{}

		var session interface{}
		if session = this.GetSession("WOKUID"); session == nil {
			return false, "未登录"
		}

		if _ok := member.FindOne(session.(string)); !_ok {
			return false, "用户不存在"
		}

		topic := &models.Topic{}
		if !topic.FindById(this.GetString("topic")) {
			return false, "文章不存在"
		}

		game := &models.Game{}
		if _ok := game.FindPath(topic.Game); !_ok {
			return false, "板块不存在"
		}

		// 需要权限为作者或管理组
		if member.Id != game.Manager && member.Id != topic.Author && !inArray(member.Id.Hex(), game.Managers) {
			return false, "没有权限"
		}

		// tag减少
		tag := &models.Tag{}
		tag.CountReduce(game.Id, this.GetString("name"))

		// 文章删除tag
		topic.RemoveTag(this.GetString("name"))

		return true, ""
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 搜索标签
func (this *TagController) SearchTag() {
	tag := &models.Tag{}
	results := tag.Like(this.GetString("game"), this.GetString("query"))

	suggestions := make([]string, len(results))

	for k, _ := range results {
		suggestions[k] = results[k].Name
	}

	this.Data["json"] = map[string]interface{}{
		"query":       "Unit",
		"suggestions": suggestions,
	}
	this.ServeJson()
}

// 获取列表
func (this *TagController) GetList() {
	ok, data := func() (bool, interface{}) {
		if len([]rune(this.GetString("name"))) > 30 {
			return false, "标签最大长度为30"
		}

		from, _ := this.GetInt("from")
		number, _ := this.GetInt("number")

		if number > 100 {
			return false, "最多显示100项"
		}

		//查找标签
		topic := &models.Topic{}
		articles := topic.FindByTag(this.GetString("game"), this.GetString("tag"), from, number)

		//找出前3个图片并取内容前200个字符
		images := make([][]string, len(articles))
		for k, _ := range articles {
			//查询每个文章前3个图片
			images[k] = FindImages(articles[k].Content, 3)
			//内容长度截取
			articles[k].Content = beego.Substr(articles[k].Content, 0, 200)
		}

		//查询该分类文章总数
		count := topic.FindTagCount(this.GetString("game"), this.GetString("tag"))

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

// 获得排名前30的标签
func (this *TagController) Hot() {
	ok, data := func() (bool, interface{}) {

		tag := &models.Tag{}
		result := tag.Hot(this.GetString("game"))

		return true, result
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 获取具有相同标签的文章
func (this *TagController) Same() {
	ok, data := func() (bool, interface{}) {
		// 查找文章
		topic := &models.Topic{}
		if _ok := topic.FindOne(this.GetString("game"), this.GetString("id")); !_ok {
			return false, "文章不存在"
		}

		result := topic.Same()

		//内容截取长度20
		for k, _ := range result {
			result[k].Content = beego.Substr(result[k].Content, 0, 20)
		}

		return true, result
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}
