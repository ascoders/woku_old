package controllers

import (
	"regexp"
	"woku/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
)

type ArticleController struct {
	beego.Controller
}

//文章返回值具体信息
type ArticleReturn struct {
	Art          *models.Article //文章指针
	Image        string          //首个图片地址，为空则不显示图片
	AuthorName   string          //作者名称
	CategoryName string          //所属分类名称
	CategoryPath string          //所属分类路径
	Time         string          //标准化时间
}

/* 分类显示页面 */
func (this *ArticleController) Category() {
	//是否用手机访问
	isMobile := IsMobile(this.Ctx.Input.UserAgent())
	//验证
	valid := validation.Validation{}
	valid.Alpha(this.Ctx.Input.Param(":category"), "category")
	if valid.HasErrors() { //没有通过验证则返回首页
		this.Ctx.Redirect(302, this.UrlFor("IndexController.Index"))
		this.StopRun()
	}

	//获得页面数
	page, _ := this.GetInt("page")

	if page < 1 {
		page = 1
	}

	category := &models.Category{}
	ok, parent, childs, articles, hots, allPage := category.FindByName(this.Ctx.Input.Param(":category"), (page-1)*15, page*15, 15)
	if !ok { //不存在此分类
		this.Ctx.Redirect(302, this.UrlFor("IndexController.Index"))
		this.StopRun()
	}
	//保存分类信息
	mainCategory := *category
	results := make([]ArticleReturn, len(articles))
	//查询每个文章的分类信息
	for k, _ := range articles {
		category.FindOne(articles[k].Category)
		results[k].Art = articles[k]
		//查找第一个图片
		results[k].Image = FindImages(results[k].Art.Content, 1)[0]
		//移除markdown标识
		results[k].Art.Content = RemoveMarkdown(results[k].Art.Content)
		//截取长度200
		results[k].Art.Content = beego.Substr(results[k].Art.Content, 0, 200)
		results[k].CategoryName = category.Title
		results[k].CategoryPath = category.Name
		results[k].Time = results[k].Art.Time.Format("2006-01-02 15:04:05")
	}
	//查询每个文章作者信息
	member := &models.Member{}
	for k, _ := range articles {
		member.FindOne(articles[k].Uid.Hex())
		results[k].AuthorName = member.Nickname
	}
	//赋值模版变量
	this.Data["category"] = mainCategory
	this.Data["parent"] = parent
	this.Data["childs"] = childs
	this.Data["articles"] = results
	this.Data["hots"] = hots
	this.Data["allpage"] = allPage
	this.Data["page"] = page //当前页数
	if isMobile {            //如果是手机访问
		this.TplNames = "mobile/article/category.html"
	} else {
		this.TplNames = "article/category.html"
	}
	this.Render()
}

/* 获取分类部分文章内容 */
func (this *ArticleController) CategoryGetArticles() {
	category := &models.Category{}
	//获得页面数
	page, _ := this.GetInt("page")

	if page < 1 {
		page = 1
	}

	ok, _, _, articles, _, _ := category.FindByName(this.GetString("category"), int(page-1)*15, int(page)*15, 15)

	if !ok { //不存在此分类
		this.StopRun()
	}

	results := make([]ArticleReturn, len(articles))

	//查询每个文章的分类信息
	for k, _ := range articles {
		category.FindOne(articles[k].Category)
		results[k].Art = articles[k]
		//查找第一个图片
		results[k].Image = FindImages(results[k].Art.Content, 1)[0]
		//移除markdown标识
		results[k].Art.Content = RemoveMarkdown(results[k].Art.Content)
		//截取长度200
		if isMobile := IsMobile(this.Ctx.Input.UserAgent()); isMobile {
			results[k].Art.Content = beego.Substr(results[k].Art.Content, 0, 100)
		} else {
			results[k].Art.Content = beego.Substr(results[k].Art.Content, 0, 200)
		}
		results[k].CategoryName = category.Title
		results[k].CategoryPath = category.Name
		results[k].Time = results[k].Art.Time.Format("2006-01-02 15:04:05")
	}

	//查询每个文章作者信息和日期
	member := &models.Member{}
	for k, _ := range articles {
		member.FindOne(articles[k].Uid.Hex())
		results[k].AuthorName = member.Nickname
	}

	this.Data["json"] = results
	this.ServeJson()
}

/* 文章界面 */
func (this *ArticleController) Article() {
	//是否用手机访问
	isMobile := IsMobile(this.Ctx.Input.UserAgent())

	//验证
	valid := validation.Validation{}
	valid.Match(this.Ctx.Input.Param(":id"), regexp.MustCompile("^[a-z0-9]+$"), "category")
	if valid.HasErrors() { //没有通过验证则返回首页
		this.Ctx.Redirect(302, this.UrlFor("IndexController.Index"))
		this.StopRun()
	}

	//寻找文章
	article := &models.Article{}
	ok := article.FindOne(this.Ctx.Input.Param(":id"))
	if !ok { //文章不存在
		this.Ctx.Redirect(302, this.UrlFor("IndexController.Index"))
		this.StopRun()
	}
	this.Data["article"] = article

	//根据文章存储的分类id查询分类信息
	category := &models.Category{}
	category.FindOne(article.Category)
	this.Data["category"] = category

	//根据分类信息查询父分类信息以及自己分类的文章（必定存在父级分类）
	parent, articles, hots := category.FindParentAndArticles()
	this.Data["parent"] = parent
	this.Data["articles"] = articles
	this.Data["hots"] = hots

	if isMobile { //如果是手机访问
		this.TplNames = "mobile/article/article.html"
	} else {
		this.TplNames = "article/article.html"
	}
	this.Render()
}

/* 文章新增浏览数post */
func (this *ArticleController) AddViews() {
	article := &models.Article{}
	article.AddViews(this.GetString("id"))
}
