package routers

import (
	"woku/controllers"
	"woku/controllers/yuqing"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/dchest/captcha"
)

func init() {
	/* 过滤器 */
	//全部请求防止恶意代理
	var FilterProxy = func(ctx *context.Context) {
		if controllers.IsProxy(ctx.Input.Domain()) { //防止恶意代理
			ctx.Redirect(302, "/")
		}
	}
	beego.InsertFilter("/*", beego.BeforeRouter, FilterProxy)

	//全局入口
	beego.Router("/", &controllers.IndexController{}, "get:Global")
	beego.Router("/*", &controllers.IndexController{}, "get:Global")

	//api
	beego.AddNamespace(beego.NewNamespace("/api",
		//首页模块
		beego.NSNamespace("/index",
			//获取首页内容
			beego.NSRouter("/getContent", &controllers.IndexController{}, "post:GetContent"),
		),
		//验证模块
		beego.NSNamespace("/check",
			//登陆
			beego.NSRouter("/login", &controllers.CheckController{}, "post:Login"),
			//注册
			beego.NSRouter("/register", &controllers.CheckController{}, "post:Register"),
			//注销
			beego.NSRouter("/signout", &controllers.CheckController{}, "post:Signout"),
			//根据md5 token自动登陆
			beego.NSRouter("/auth", &controllers.CheckController{}, "post:Auth"),
			//社交化平台查询是否有账号，若有自动登陆
			beego.NSRouter("/hasOauth", &controllers.CheckController{}, "post:HasOauth"),
			//第三方平台注册用户
			beego.NSRouter("/oauthRegister", &controllers.CheckController{}, "post:OauthRegister"),
		),
		//游戏模块
		beego.NSNamespace("/game",
			//获取所有游戏列表
			beego.NSRouter("/getGameList", &controllers.GameController{}, "post:GetGameList"),
			//创建游戏
			beego.NSRouter("/createGame", &controllers.GameController{}, "post:CreateGame"),
			//获取游戏基础信息
			beego.NSRouter("/getInfo", &controllers.GameController{}, "post:GetInfo"),
			//获取x游戏y分类的列表
			beego.NSRouter("/getList", &controllers.GameController{}, "post:GetList"),
			//获取x游戏文章信息
			beego.NSRouter("/getPage", &controllers.GameController{}, "post:GetPage"),
			//发帖/回帖/嵌套回复
			beego.NSRouter("/addTopic", &controllers.GameController{}, "post:AddTopic"),
			//将帖子移动到其他分类下
			beego.NSRouter("/changeCategory", &controllers.GameController{}, "post:ChangeCategory"),
			//对帖子操作 置顶/加精/删除
			beego.NSRouter("/operate", &controllers.GameController{}, "post:Operate"),
			//上传图片后处理（与七牛服务器交互）
			beego.NSRouter("/uploadHandle", &controllers.GameController{}, "post:UploadHandle"),
			//管理 基本信息保存
			beego.NSRouter("/baseSave", &controllers.GameController{}, "post:BaseSave"),
			//管理 分类管理 修改优先级
			beego.NSRouter("/changeRecommendPri", &controllers.GameController{}, "post:ChangeRecommendPri"),
			//管理 新增分类
			beego.NSRouter("/addCategory", &controllers.GameController{}, "post:AddCategory"),
			//管理 更新分类
			beego.NSRouter("/updateCategory", &controllers.GameController{}, "post:UpdateCategory"),
			//管理 删除分类
			beego.NSRouter("/deleteCategory", &controllers.GameController{}, "post:DeleteCategory"),
		),
		//用户后台模块
		beego.NSNamespace("/user",
			//获取消息列表
			beego.NSRouter("/getMessage", &controllers.UserController{}, "post:GetMessage"),
			//获取充值记录
			beego.NSRouter("/getHistory", &controllers.UserController{}, "post:GetHistory"),
			//修改密码
			beego.NSRouter("/password", &controllers.UserController{}, "post:Password"),
			//生成付款订单并返回跳转表单
			beego.NSRouter("/recharge", &controllers.UserController{}, "post:Recharge"),
			//为已有未过期、未完成订单继续付款
			beego.NSRouter("/rechargeOrder", &controllers.UserController{}, "post:RechargeOrder"),
			//修改绑定邮箱 - 发送邮件
			beego.NSRouter("/email", &controllers.UserController{}, "post:Email"),
			//获取第三方平台绑定状况列表
			beego.NSRouter("/oauthList", &controllers.UserController{}, "post:OauthList"),
			//第三方平台绑定 新增/更新
			beego.NSRouter("/oauth", &controllers.UserController{}, "post:Oauth"),
			//修改用户头像
			beego.NSRouter("/changeImage", &controllers.UserController{}, "post:ChangeImage"),
			//账号管理（需要最高权限）
			beego.NSRouter("/member", &controllers.UserController{}, "post:Member"),
			//查询有哪些职位（需要最高权限）
			beego.NSRouter("/jobs", &controllers.UserController{}, "post:Jobs"),
			//职位管理（需要最高权限）
			beego.NSRouter("/job", &controllers.UserController{}, "post:Job"),
		),
		//七牛图片处理模块
		beego.NSNamespace("/qiniu",
			//获取首页内容
			beego.NSRouter("/createUpToken", &controllers.QiniuController{}, "post:CreateUpToken"),
		),
		//标签
		beego.NSNamespace("/tag",
			//绑定标签
			beego.NSRouter("/bind", &controllers.TagController{}, "post:Bind"),
			//解绑标签
			beego.NSRouter("/unBind", &controllers.TagController{}, "post:UnBind"),
			//提示推荐标签
			beego.NSRouter("/searchTag", &controllers.TagController{}, "post:SearchTag"),
			//获取标签列表
			beego.NSRouter("/getList", &controllers.TagController{}, "post:GetList"),
			//获取前30个热门标签
			beego.NSRouter("/hot", &controllers.TagController{}, "post:Hot"),
			//相似标签
			beego.NSRouter("/same", &controllers.TagController{}, "post:Same"),
		),
		// 文档
		beego.NSNamespace("/doc",
			// 获取文档
			beego.NSRouter("/getDoc", &controllers.DocController{}, "post:GetDoc"),
			// 删除文件夹
			beego.NSRouter("/deleteFolder", &controllers.DocController{}, "post:DeleteFolder"),
			// 根据文章id查询之前节点信息
			beego.NSRouter("/parents", &controllers.DocController{}, "post:Parents"),
			// 更新文档排序
			beego.NSRouter("/exchange", &controllers.DocController{}, "post:Exchange"),
		),
		// 舆情分析
		beego.NSNamespace("/yuqing",
			// 分词管理
			beego.NSRouter("/split", &controllers.YuqingController{}, "post:Split"),
			// 分析管理
			beego.NSRouter("/analyse", &controllers.YuqingController{}, "post:Analyse"),
			// 舆情列表
			beego.NSRouter("/", &controllers.YuqingController{}, "post:Index"),
			// 载入分词词库
			beego.NSRouter("/loadSego", &controllers.YuqingController{}, "post:LoadSego"),
			// 载入分析词库
			beego.NSRouter("/loadAnaylse", &controllers.YuqingController{}, "post:LoadAnaylse"),
			// 抓取信息
			beego.NSRouter("/getRss", &controllers.YuqingController{}, "post:GetRss"),
			// 抓取信息
			beego.NSRouter("/freshResult", &controllers.YuqingController{}, "post:FreshResult"),
			// 载入、抓取等操作完成状态
			beego.NSRouter("/operateStatus", &controllers.YuqingController{}, "post:OperateStatus"),
			// 图标信息
			beego.NSRouter("/charts", &controllers.YuqingController{}, "post:Charts"),
			// 中国今天的舆情
			beego.NSRouter("/china", &yuqing.China{}, "post:Read"),
			// 舆情列表/分类
			beego.NSRouter("/:category", &controllers.YuqingController{}, "post:Index"),
		),
		// 获取登陆的用户信息
		beego.NSRouter("/currentUser", &controllers.CheckController{}, "post:CurrentUser"),
		// 刷新验证码
		beego.NSRouter("/freshCap", &controllers.CheckController{}, "post:FreshCap"),
	))

	beego.Handler("/captcha/*.png", captcha.Server(240, 80)) //获取验证码图片 240 x 80

	/*

		//手机支付页面
		beego.Router("/mobile/pay.html", &controllers.ApiController{}, "get:MobilePay")

		/* 注册登录 */
	/*
		//登陆页面、登陆处理页面
		beego.Router("/login.html", &controllers.CheckController{}, "get:Login;post:PostLogin")
		//注册页面
		beego.Router("/register.html", &controllers.CheckController{}, "get:Register")

		/* check */
	/*
		beego.AddNamespace(beego.NewNamespace("/check",
			beego.NSRouter("/findpass.html", &controllers.CheckController{}, "get:FindPass;post:FindPassPost"), //找回密码页面
			beego.NSRouter("/jump.html", &controllers.CheckController{}, "get:Jump"),                           //授权码转发页面
			beego.NSRouter("/notify.html", &controllers.CheckController{}, "get:Notify"),                       //第三方登陆回调页面
			beego.NSNamespace("/post",
				beego.NSRouter("/notify", &controllers.CheckController{}, "post:NotifyPost"),             //第三方登陆回调页面-post
				beego.NSRouter("/notifyRegister", &controllers.CheckController{}, "post:NotifyRegister"), //第三方登陆，自动注册账号-post
				beego.NSRouter("/current_user", &controllers.CheckController{}, "post:CurrentUser"),      //检测登陆状态
			),
		))

		/* 网站接口 */
	/*
		beego.Router("/user.html", &controllers.UserController{}, "get:Index") //用户首页
		beego.AddNamespace(beego.NewNamespace("/web",
			beego.NSNamespace("/admin",
				beego.NSRouter("/account_base", &controllers.UserController{}, "post:AccountBase"),                           //基本信息
				beego.NSRouter("/account_message", &controllers.UserController{}, "post:AccountMessage"),                     //消息中心
				beego.NSRouter("/account_messagepost", &controllers.UserController{}, "post:AccountMessagePost"),             //消息中心-post分页查询
				beego.NSRouter("/account_messagereadpost", &controllers.UserController{}, "post:AccountMessageReadPost"),     //消息中心-post已读请求
				beego.NSRouter("/account_messagenumberpost", &controllers.UserController{}, "post:AccountMessageNumber"),     //消息中心-post获取未读消息数量
				beego.NSRouter("/account_reward", &controllers.UserController{}, "post:AccountReward"),                       //奖励额度
				beego.NSRouter("/account_history", &controllers.UserController{}, "post:AccountHistory"),                     //充值记录
				beego.NSRouter("/account_historypost", &controllers.UserController{}, "post:AccountHistoryPost"),             //充值记录-post查询
				beego.NSRouter("/account_changepass", &controllers.UserController{}, "post:AccountChangepass"),               //修改密码
				beego.NSRouter("/account_changepasspost", &controllers.UserController{}, "post:AccountChangepassPost"),       //修改密码-post
				beego.NSRouter("/account_recharge", &controllers.UserController{}, "post:AccountRecharge"),                   //账户充值
				beego.NSRouter("/account_rechargepost", &controllers.UserController{}, "post:AccountRechargePost"),           //账户充值-post 生成新账单
				beego.NSRouter("/account_rechargeorderpost", &controllers.UserController{}, "post:AccountRechargeOrderPost"), //账户充值-post 对已有账单处理
				beego.NSRouter("/account_other", &controllers.UserController{}, "post:AccountOther"),                         //账户绑定第三方平台
				beego.NSRouter("/account_free", &controllers.UserController{}, "post:AccountFree"),                           //代金券
				beego.NSRouter("/account_email", &controllers.UserController{}, "post:AccountEmail"),                         //修改绑定邮箱
				beego.NSRouter("/account_emailpost", &controllers.UserController{}, "post:AccountEmailPost"),                 //修改绑定邮箱-post
				beego.NSRouter("/job_money", &controllers.UserController{}, "post:JobMoney"),                                 //领取工资
				beego.NSRouter("/job_promotion", &controllers.UserController{}, "post:JobPromotion"),                         //我要升职
				beego.NSRouter("/job_salary", &controllers.UserController{}, "post:JobSalary"),                               //薪资管理
				beego.NSRouter("/job_salarypost", &controllers.UserController{}, "post:JobSalaryPost"),                       //薪资管理post
				beego.NSRouter("/job_manage", &controllers.UserController{}, "post:JobManage"),                               //人员管理
				beego.NSRouter("/job_managepostlist", &controllers.UserController{}, "post:JobManagePostList"),               //人员管理获取列表post
				beego.NSRouter("/job_managepost", &controllers.UserController{}, "post:JobManagePost"),                       //人员管理post
				beego.NSRouter("/job_manage_finduser", &controllers.UserController{}, "get:JobManageFindUser"),               //搜索用户
				beego.NSRouter("/article_new", &controllers.UserController{}, "post:ArticleNew"),                             //发布新文章
				beego.NSRouter("/article_newpost", &controllers.UserController{}, "post:ArticleNewPost"),                     //新文章提交post
				beego.NSRouter("/article_mylist", &controllers.UserController{}, "post:MyArticleList"),                       //我的文章列表
				beego.NSRouter("/article_mylistpage", &controllers.UserController{}, "post:MyArticleListPage"),               //我的文章列表请求分页内容
				beego.NSRouter("/article_mylistpost", &controllers.UserController{}, "post:MyArticleListPost"),               //我的文章列表post
				beego.NSRouter("/article_get_content", &controllers.UserController{}, "post:GetArticleContent"),              //获取文章信息
				beego.NSRouter("/article_category", &controllers.UserController{}, "post:ArticleCategory"),                   //分类管理
				beego.NSRouter("/article_category_post", &controllers.UserController{}, "post:ArticleCategoryPost"),          //分类管理post操作
				beego.NSRouter("/game_category", &controllers.UserController{}, "post:GameCategory"),                         //讨论组管理
				beego.NSRouter("/game_category_post", &controllers.UserController{}, "post:GameCategoryPost"),                //讨论组修改post
				beego.NSRouter("/web_count", &controllers.UserController{}, "post:WebCount"),                                 //网站统计
				beego.NSRouter("/web_useraction", &controllers.UserController{}, "post:WebUserAction"),                       //用户行为
				beego.NSRouter("/situation_spider", &controllers.UserController{}, "post:SituationSpider"),                   //抓取地址管理
				beego.NSRouter("/situation_spider_operate", &controllers.UserController{}, "post:SituationSpiderOperate"),    //抓取地址管理 增删改查
				beego.NSRouter("/situation_text", &controllers.UserController{}, "post:SituationText"),                       //正则寻址管理
				beego.NSRouter("/situation_text_operate", &controllers.UserController{}, "post:SituationTextOperate"),        //正则寻址管理 增删改查
				beego.NSRouter("/situation_text_like", &controllers.UserController{}, "post:SituationTextSelectLike"),        //正则寻址管理 模糊匹配
				beego.NSRouter("/situation_error", &controllers.UserController{}, "post:SituationError"),                     //抓取错误
				beego.NSRouter("/situation_error_operate", &controllers.UserController{}, "post:SituationErrorOperate"),      //抓取错误 增删改查
				beego.NSRouter("/user_upload", &controllers.UserController{}, "post:UserUpload"),                             //后台用户上传文件
				beego.NSRouter("/user_clear_message", &controllers.UserController{}, "post:ClearMessage"),                    //清空消息
				beego.NSRouter("/user_bind_notify", &controllers.UserController{}, "get:BindNotify"),                         //绑定第三方平台回调页面
				beego.NSRouter("/user_add_social", &controllers.UserController{}, "post:AddSocial"),                          //新增第三方平台
				beego.NSRouter("/user_change_user_head_image", &controllers.UserController{}, "post:ChangeUserHeadImage"),    //修改用户头像
			),
			beego.NSNamespace("/api", //网站api
				beego.NSRouter("/hassession", &controllers.CheckController{}, "post:HasSession"), //验证是否有登陆
				beego.NSRouter("/register", &controllers.CheckController{}, "post:PostRegister"), //注册处理页面
				beego.NSRouter("/exist", &controllers.CheckController{}, "post:DestroySession"),  //退出登陆
				beego.NSRouter("/checklogin", &controllers.CheckController{}, "post:CheckLogin"), //验证是否登陆
				beego.NSRouter("/autologin", &controllers.CheckController{}, "post:AutoLogin"),   //自动登陆
			),
		))

		/* 文章 */
	/*
		beego.AddNamespace(beego.NewNamespace("/article",
			beego.NSRouter("/:id([0-9a-z]+).html", &controllers.ArticleController{}, "get:Article"), //文章页面
			beego.NSNamespace("/post",
				beego.NSRouter("/addViews", &controllers.ArticleController{}, "post:AddViews"), //文章新增浏览数提交页面
			),
		))

		/* 分类 */
	/*
		beego.AddNamespace(beego.NewNamespace("/category",
			beego.NSRouter("/:category([a-z]+).html", &controllers.ArticleController{}, "get:Category"),               //分类首页
			beego.NSRouter("/:category([a-z]+).html/:page([0-9]+)", &controllers.ArticleController{}, "get:Category"), //分类分页面
			beego.NSNamespace("/post",
				beego.NSRouter("/getarticles", &controllers.ArticleController{}, "post:CategoryGetArticles"), //获取分类某页的文章信息post
			),
		))

		/* 游戏 game/group */
	/*
		beego.Router("/game", &controllers.GameController{}, "get:Index;post:IndexPost") //首页/获取列表
		beego.Router("/game/*", &controllers.GameController{}, "get:Index")              //列表页
		beego.AddNamespace(beego.NewNamespace("/g",
			beego.NSRouter("/:game([a-z]+)", &controllers.GameController{}, "get:Category;post:CategoryPost"),                        //游戏各分类首页和帖子内容页
			beego.NSRouter("/:game([a-z]+)/:category([0-9a-z]+)", &controllers.GameController{}, "get:Category;post:CategoryPost"),   //游戏各分类首页和帖子内容页
			beego.NSRouter("/:game([a-z]+)/:category([0-9a-z]+)/*", &controllers.GameController{}, "get:Category;post:CategoryPost"), //游戏各分类首页和帖子内容页
			beego.NSNamespace("/post",
				     //创建游戏请求
				beego.NSRouter("/getlist", &controllers.GameController{}, "post:GetList"),           //获取列表页
				beego.NSRouter("/getdetail", &controllers.GameController{}, "post:GetDetail"),       //获取详细页
				beego.NSRouter("/addtopic", &controllers.GameController{}, "post:AddTopic"),         //发帖&回帖请求post
				beego.NSRouter("/topicoperate", &controllers.GameController{}, "post:TopicOperate"), //话题置顶之类操作post
				beego.NSRouter("/findreply", &controllers.GameController{}, "post:FindReply"),       //查询子评论post
				beego.NSRouter("/upload", &controllers.GameController{}, "post:Upload"),             //上传
				beego.NSRouter("/basesave", &controllers.GameController{}, "post:BaseSave"),         //管理-基本息保存
			),
		))

		/* API接口 */
	/*
		beego.AddNamespace(beego.NewNamespace("/api",
			beego.NSRouter("/", &controllers.ApiController{}, "post:Api;get:Api"),                              //api统一接口
			beego.NSRouter("/alipayreturn.html", &controllers.ApiController{}, "get:AlipayReturn"),             //接收用户在支付宝支付成功后跳转回的页面
			beego.NSRouter("/alipaymobilereturn.html", &controllers.ApiController{}, "get:AlipayMobileReturn"), //接收用户在手机网页版支付宝支付成功后跳转回的页面
			beego.NSRouter("/alipaynotify", &controllers.ApiController{}, "post:AlipayNotify"),                 //接收用户在支付宝支付成功后异步通知地址
		))
	*/

	/* 单个页面 */
	/*
		beego.AddNamespace(beego.NewNamespace("/page",
			beego.NSRouter("/solution.html", &controllers.PageController{}, "get:Solution"),      //解决方案
			beego.NSRouter("/api.html", &controllers.PageController{}, "get:Api"),                //开发文档
			beego.NSRouter("/code_format.html", &controllers.PageController{}, "get:CodeFormat"), //代码格式化
		))

		/* socket */
	/*
		beego.Router("/ws/socket", &controllers.SocketController{}, "get:Socket")

		/* 舆情分析 */
	/*
		beego.Router("/situation.html", &controllers.SituationController{}, "get:Index") //首页
		beego.AddNamespace(beego.NewNamespace("/situation"))

		/* 手机 */
	/*
		beego.AddNamespace(beego.NewNamespace("/mobile",
			beego.NSNamespace("/post",
				beego.NSRouter("/category", &controllers.ArticleController{}, "post:CategoryGetArticles"), //获取分类列表
			),
		))

		/* 验证码 */
	/*
		beego.Handler("/captcha/*.png", captcha.Server(240, 80)) //注册验证码服务，验证码图片的宽高为240 x 80
	*/
}
