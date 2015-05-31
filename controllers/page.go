package controllers

import (
	"github.com/astaxie/beego"
)

type PageController struct {
	beego.Controller
}

/* 解决方案 */
func (this *PageController) Solution() {
	this.TplNames = "page/solution.html"
	hasSession := false
	if this.GetSession("WOKUID") != nil {
		hasSession = true
	}
	this.Data["session"] = hasSession
	this.Render()
}

/* api文档 */
func (this *PageController) Api() {
	this.TplNames = "page/api.html"
	this.Render()
}

/* 代码格式化 */
func (this *PageController) CodeFormat() {
	this.TplNames = "page/code_format.html"
	this.Render()
}
