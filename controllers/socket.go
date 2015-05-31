package controllers

import (
	"os"
	"strconv"
	"woku/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
)

type SocketController struct {
	beego.Controller
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	SocketLog *logs.BeeLogger            //打印日志
	Clients   map[string]*websocket.Conn //客户端队列
)

func init() {
	//如果没有日志目录则创建日志目录
	_, err := os.Open("log")
	if err != nil && os.IsNotExist(err) {
		os.Mkdir("log", 0777)
	}
	//初始化日志
	SocketLog = logs.NewLogger(10000)
	SocketLog.SetLogger("file", `{"filename":"log/socket.log"}`)
	//初始化客户端
	Clients = make(map[string]*websocket.Conn, 1000)
}

func (this *SocketController) Socket() {
	conn, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
	//如果没有session则退出
	id := this.GetSession("WOKUID")
	if id == nil { //用户不存在
		conn.Close()
		this.StopRun()
	} else { //用户存在
		//判断该链接是否存在
		if _, ok := Clients[id.(string)]; !ok { //链接不存在，表示用户只有一个浏览器标签访问
			//查询用户信息
			member := &models.Member{}
			member.FindOne(id.(string))
			conn.WriteMessage(1, []byte(strconv.Itoa(member.MessageNumber)))
			defer delete(Clients, id.(string))
		}
		//保存这个用户的链接
		Clients[id.(string)] = conn
	}
	defer func() {
		conn.Close()
	}()
	if err != nil {
		SocketLog.Error("Upgrade:", err)
		this.StopRun()
	}
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
