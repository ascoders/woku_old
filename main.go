package main

import (
	"woku/controllers"
	_ "woku/controllers/yuqing"
	_ "woku/routers"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
)

func main() {
	// 定期执行计划任务
	go controllers.PlanTask()

	// 运行
	beego.Run()
}
