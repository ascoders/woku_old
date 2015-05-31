package controllers

import (
	"woku/models"

	"github.com/astaxie/beego"
)

type IndexController struct {
	beego.Controller
}

func (this *IndexController) Global() {
	this.TplNames = "html/public/global.html"
	this.Render()
}

func (this *IndexController) GetContent() {
	//获取最新发布的资讯
	article := &models.Article{}
	articles := article.FindNews()

	//查询最火的前12款游戏
	game := &models.Game{}
	games := game.FindHot(12)

	//查询最新的前10款游戏
	newGames := game.FindGame(0, 10)

	//查找本周评论数最多的前10个帖子
	topic := &models.Topic{}
	hotTopics := topic.WeekTopBbs()

	this.Data["json"] = map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"Tops":      articles,
			"Games":     games,
			"NewGames":  newGames,
			"HotTopics": hotTopics,
		},
	}

	this.ServeJson()
}
