'use strict';

///////////////////////////////////////////////////////////////////////////////////////////////
// avalon
///////////////////////////////////////////////////////////////////////////////////////////////

//改变模板标签
avalon.config({
	interpolate: ["{[{", "}]}"]
});

//过滤markdown标签
avalon.filters.cleanmark = function (str) {
	//移除所有 * ` [ ] # - >
	str = str
		.replace(/[!.*](.*)/g, "【图片】")
		.replace(/\*/g, "")
		.replace(/\`/g, "")
		.replace(/\[/g, "")
		.replace(/\]/g, "")
		.replace(/\#/g, "")
		.replace(/\-/g, "")
		.replace(/\>/g, "");

	return str;
};

//处理小数点
avalon.filters.toFixed = function (str, number) {
	str = str.toFixed(number);

	return str;
};

require.config({
	baseUrl: "/static/",
	paths: {
		"jquery": "http://cdn.bootcss.com/jquery/1.11.2/jquery.min",
		"jQuery": "http://cdn.bootcss.com/jquery/1.11.2/jquery.min",
		"jquery.timeago": "http://cdn.bootcss.com/jquery-timeago/1.4.0/jquery.timeago.min", //友好时间
		"jquery.ui": "http://cdn.bootcss.com/jqueryui/1.10.4/jquery-ui.min", //jquery-ui
		"jquery.autosize": "http://cdn.bootcss.com/autosize.js/1.18.15/jquery.autosize.min", //textarea大小自适应高度
		"jquery.selection": "http://cdn.bootcss.com/jquery.selection/1.0.1/jquery.selection.min", //表单选择
		"jquery.qrcode": "http://cdn.bootcss.com/jquery.qrcode/1.0/jquery.qrcode.min", //二维码
		"jquery.cookie": "http://cdn.bootcss.com/jquery-cookie/1.4.1/jquery.cookie", //操作cookie
		"jquery.autocomplete": "http://cdn.bootcss.com/jquery.devbridge-autocomplete/1.2.7/jquery.devbridge-autocomplete.min", //输入框自动补全
		"dropzone": "http://cdn.bootcss.com/dropzone/3.12.0/dropzone-amd-module.min", //拖拽上传
		"prettify": "http://cdn.bootcss.com/prettify/r298/prettify.min", //code美化
		// "chart": "http://cdn.bootcss.com/Chart.js/1.0.1-beta.2/Chart.min", //表格
		"md5": "http://cdn.bootcss.com/blueimp-md5/1.1.0/js/md5.min", //md5加密
		"echarts": 'http://cdn.bootcss.com/echarts/2.1.10/echarts-all', // 百度表格

		"mmState": 'js/plugin/mmState',
		"mmRouter": 'js/plugin/mmRouter',
		"mmHistory": 'js/plugin/mmHistory',
		"mmPromise": 'js/plugin/mmPromise',
		"jquery.tree": "js/plugin/jquery.treeview", //树状菜单
		"jquery.typetype": "js/plugin/jquery.typetype", //模拟输入
		"jquery.taboverride": "js/plugin/taboverride", //tab键变为缩进
		"jquery.contextMenu": "js/plugin/jquery.contextMenu", //右键菜单
		"jquery.jbox": "js/plugin/jBox", //迷你提示框
		"marked": "js/plugin/marked", //markdown解析
		"frontia": "js/plugin/baidu.frontia.1.0.0", //百度社会化组件

		"editor": "js/public/editor", //编辑器
		"avalon.table": "js/public/avalon.table", // 超级表格
		"avalon.page": "js/public/avalon.page" //分页
	},
	shim: {
		'jquery.timeago': {
			deps: ['jquery']
		},
		'jquery.ui': {
			deps: ['jquery']
		},
		'jquery.jbox': {
			deps: ['jquery']
		},
		'jquery.autosize': {
			deps: ['jquery']
		},
		'jquery.taboverride': {
			deps: ['jquery']
		},
		'jquery.selection': {
			deps: ['jquery']
		},
		'jquery.qrcode': {
			deps: ['jquery']
		},
		'jquery.typetype': {
			deps: ['jquery']
		},
		'jquery.autocomplete': {
			deps: ['jquery']
		},
		'jquery.tree': {
			deps: ['jquery']
		},
		'md5': {
			exports: 'md5'
		},
		'frontia': {
			exports: 'baidu.frontia'
		},
		'jquery.contextMenu': {
			deps: ['jquery']
		},
		'echarts': {
			exports: "echarts"
		}
	}
});

///////////////////////////////////////////////////////////////////////////////////////////////
// 自定义函数
///////////////////////////////////////////////////////////////////////////////////////////////

//封装提示框
function notice(text, color) {
	require(['jquery', 'jquery.jbox'], function ($) {
		new jBox('Notice', {
			content: text,
			attributes: {
				x: 'right',
				y: 'bottom'
			},
			animation: 'flip',
			color: color
		});
	});
}

//封装选择框
function confirm(content, callback) {
	require(['jquery', 'jquery.jbox'], function ($) {
		var myModal = new jBox('Confirm', {
			minWidth: '200px',
			content: content,
			animation: 'flip',
			confirmButton: '确定',
			cancelButton: '取消',
			confirm: function () {
				callback();
			}
		});

		myModal.open();
	});
}

//调整用户头像图片路径
function userImage(str) {
	if (str != undefined && str != "") {
		if (!isNaN(str)) {
			return "/static/img/user/" + str + ".jpg";
		}
		return str;
	}
	return;
}

//自带_xsrf的post提交，整合了error处理
function post(url, params, success, error, callback, errorback) {
	require(['jquery', 'jquery.cookie'], function ($) {
		//获取xsrftoken
		var xsrf = $.cookie("_xsrf");
		if (!xsrf) {
			return;
		}
		var xsrflist = xsrf.split("|");
		var xsrftoken = Base64.decode(xsrflist[0]);

		var postParam = {
			_xsrf: xsrftoken
		};
		postParam = $.extend(postParam, params);

		return $.ajax({
				url: url,
				type: 'POST',
				traditional: true, //为了传数组
				data: postParam
			})
			.done(function (data) {
				if (data.ok) { //操作成功
					if (success !== null) {
						notice(success, 'green');
					}
					//执行回调函数
					if (callback != null) {
						callback(data.data);
					}
				} else { //操作失败
					if (error !== null) {
						notice(error + data.data, 'red');
					}
					if (errorback != null) {
						errorback(data.data);
					}
				}
			});
	});
}

//创建分页DOM内容
function createPagin(from, number, count, params) {
	//计算总页数
	if (count == 0) {
		return '';
	}
	var allPage = Math.ceil(parseFloat(count) / parseFloat(number));

	//如果总页数是1，返回空
	if (allPage == 1) {
		return '';
	}

	//当前页数
	var page = 1;

	from = parseInt(from);
	number = parseInt(number);

	if (from != 0) {
		page = from / number + 1;
	}

	//中间内容
	var list = "";

	//附加参数
	var paramString = '';
	for (var key in params) {
		paramString += '&' + key + '=' + params[key];
	}


	var path = window.location.pathname + '?number=' + number + paramString;

	//根据页数计算from
	var _from = function (i) {
		return ((i - 1) * number);
	}

	//首部箭头
	if (page == 1) {
		list += "<li><a class='disabled f-bln' href='#'><i class='fa fa-arrow-left'></i></a></li>";
	} else {
		list += "<li><a class='f-bln' href='#!" + path + "&from=" + (from - number) + "'><i class='fa fa-arrow-left'></i></a></li>";
	}

	//中间部分
	if (allPage < 7) {
		for (var i = 1; i <= allPage; i++) {
			if (i == page) {
				list += "<li><a class='active' href='#'>" + i + "</a></li>";
			} else {
				list += "<li><a href='#!" + path + "&from=" + _from(i) + "'>" + i + "</a></li>";
			}
		}
	} else {
		if (page < 6) {
			for (var i = 1; i <= 6; i++) {
				if (i == page) {
					list += "<li><a class='active' href='#'>" + i + "</a></li>";
				} else {
					list += "<li><a href='#!" + path + "&from=" + _from(i) + "'>" + i + "</a></li>";
				}
			}
			list += "<li><a class='disabled' href='#'>...</a></li>";
			list += "<li><a href='#!" + path + "&from=" + _from(allPage) + "'>" + allPage + "</a></li>";
		} else {
			list += "<li><a href='#!" + path + "&from=" + _from(1) + "'>1</a></li>";
			list += "<li><a href='#!" + path + "&from=" + _from(2) + "'>2</a></li>";
			list += "<li><a class='disabled' href='#'>...</a></li>";
			if (allPage - page < 6) {
				for (var i = allPage - 6; i <= allPage; i++) {
					if (i == page) {
						list += "<li><a class='active' href='#'>" + i + "</a></li>";
					} else {
						list += "<li><a href='#!" + path + "&from=" + _from(i) + "'>" + i + "</a></li>";
					}
				}
			} else {
				for (var i = page - 2; i <= page + 3; i++) {
					if (i == page) {
						list += "<li><a class='active' href='#'>" + i + "</a></li>";
					} else {
						list += "<li><a href='#!" + path + "&from=" + _from(i) + "'>" + i + "</a></li>";
					}
				}
				list += "<li><a class='disabled' href='#'>...</a></li>";
				list += "<li><a href='#!" + path + "&from=" + _from(allPage - 1) + "'>" + (allPage - 1) + "</a></li>";
				list += "<li><a href='#!" + path + "&from=" + _from(allPage) + "'>" + allPage + "</a></li>";
			}
		}
	}

	//末尾箭头
	if (page == allPage) {
		list += "<li><a class='disabled f-brn' href='javascript:void(0);'><i class='fa fa-arrow-right'></i></a></li>";
	} else {
		list += "<li><a class='f-brn' href='#!" + path + "&from=" + (from + number) + "'><i class='fa fa-arrow-right'></i></a></li>";
	}

	return "<ul class='g-pa f-pr'>" + list + "</ul>";
}

//字符串截取方法，支持中文
function subStr(str, start, end) {
	var _start = 0;
	for (var i = 0; i < start; i++) {
		if (escape(str.charCodeAt(i)).indexOf("%u") >= 0) {
			_start += 2;
		} else {
			_start += 1;
		}
	}
	var _end = _start;
	for (var i = start; i < end; i++) {
		if (escape(str.charCodeAt(i)).indexOf("%u") >= 0) {
			_end += 2;
		} else {
			_end += 1;
		}
	}
	var r = str.substr(_start, _end);
	return r;
}

//dropzone统一外包一层规范
function createDropzone(obj, url, params, accept, callback) {
	require(['jquery', 'dropzone', 'md5', 'jquery.jbox'], function ($, Dropzone, md5) {
		//上传框组
		var modals = {};

		//实例化dropzone
		return new Dropzone(obj, {
			url: url,
			maxFiles: 10,
			maxFilesize: 0.5,
			method: 'post',
			acceptedFiles: accept,
			autoProcessQueue: false,
			init: function () {
				//事件监听
				this.on("addedfile", function (file) {
					//实例化上传框
					modals[md5(file.name)] = new jBox('Notice', {
						attributes: {
							x: 'left',
							y: 'bottom'
						},
						title: '上传 ' + file.name + ' 中..',
						theme: 'NoticeBorder',
						color: 'black',
						animation: {
							open: 'slide:bottom',
							close: 'slide:left'
						},
						autoClose: false,
						closeOnClick: false,
						onCloseComplete: function () {
							this.destroy();
						}
					});

					var _this = this;

					//获取上传到七牛的token
					post('/api/qiniu/createUpToken', params, null, '', function (data) {
						_this.options.params['token'] = data;

						// 开始上传
						_this.processQueue();
					}, function () { //失败撤销上传框
						modals[md5(file.name)].close();
					});
				});
				this.on("thumbnail", function (file, img) { //文件内容,缩略图base64
					//如果模态框被关闭,return
					if (!modals[md5(file.name)]) {
						return;
					}

					// 给缩略图赋值
					modals[md5(file.name)].setContent('<img src="' + img + '"><br><div class="progress" style="margin:10px 0 0 0"><div class="progress-bar" id="upload' + md5(file.name) + '" style="min-width:5%;">0%</div></div><br>尺寸: ' + file.width + ' × ' + file.height + ' &nbsp;&nbsp;大小: ' + (file.size / 1000).toFixed(1) + ' Kb<br>');
				});
				this.on("error", function (file, err) {
					notice(err.toString(), 'red');

					//如果模态框被关闭,return
					if (!modals[md5(file.name)]) {
						return;
					}

					//模态框关闭
					modals[md5(file.name)].close();
					modals[md5(file.name)] = null;
				});
				this.on("uploadprogress", function (file, process, size) {
					//如果模态框被关闭,return
					if (!modals[md5(file.name)]) {
						return;
					}

					process = process.toFixed(2);

					if (process == 100) {
						process = 99;
					}

					$('#upload' + md5(file.name)).css('width', process + "%").text(process + '%');
				});
				this.on("success", function (file, data) {
					notice('上传成功', 'green');

					//如果模态框被关闭,return
					if (!modals[md5(file.name)]) {
						return;
					}

					$('#upload' + md5(file.name)).css('width', "100%").text('100%');

					setTimeout(function () {
						//如果模态框被关闭,return
						if (!modals[md5(file.name)]) {
							return;
						}
						//模态框关闭
						modals[md5(file.name)].close();
						modals[md5(file.name)] = null;
					}, 200);

					//触发回调
					callback(data, file);
				});
			}
		});
	});
}

// 判断ie9及其以下版本
function ieVersion() {
	var v = 3,
		div = document.createElement('div'),
		all = div.getElementsByTagName('i');
	while (div.innerHTML = '<!--[if gt IE ' + (++v) + ']><i></i><![endif]-->', all[0]);
	return v > 4 ? v : false;
}

// 倒计时
function timediff(element, options, callback) {
	// 初始化
	var defaults = {
		second: 0
	};
	var opts = $.extend(defaults, options);
	opts.second = parseInt(opts.second);

	function Run() {
		var day = 0,
			hour = 0,
			minute = 0,
			second = 0; //时间默认值   

		if (opts.second > 0) {
			day = Math.floor(opts.second / (60 * 60 * 24));
			hour = Math.floor(opts.second / (60 * 60)) - (day * 24);
			minute = Math.floor(opts.second / 60) - (day * 24 * 60) - (hour * 60);
			second = Math.floor(opts.second) - (day * 24 * 60 * 60) - (hour * 60 * 60) - (minute * 60);
		} else if (opts.second == 0) {
			callback();
		}
		if (minute <= 9) minute = '0' + minute;
		if (second <= 9) second = '0' + second;
		element.find("#j-day").html(day + " 天");
		element.find("#j-hour").html(hour + " 时");
		element.find("#j-minute").html(minute + " 分");
		element.find("#j-second").html(second + " 秒");
		opts.second--;
	}

	var inter = setInterval(function () {
		if (!$.contains(document, element[0])) { //dom不存在就停止事件
			clearInterval(inter);
		}
		Run();
	}, 1000);

	Run();
}

// jbox插件渲染dom
function jbox() {
	require(['jquery', 'jquery.jbox'], function ($) {
		// jbox插件
		$('.jbox').each(function () {
			var title = $(this).attr('title');
			if (!title) {
				return;
			}

			// 方向
			var jboxPositionX = $(this).attr('jbox-position-x') || 'center'
			var jboxPositionY = $(this).attr('jbox-position-y') || 'top'

			$(this).removeAttr('title');
			$(this).jBox('Tooltip', {
				content: title,
				animation: 'zoomIn',
				closeOnMouseleave: true,
				position: {
					x: jboxPositionX,
					y: jboxPositionY
				}
			});
		});
	});
}

///////////////////////////////////////////////////////////////////////////////////////////////
// 初始化插件
///////////////////////////////////////////////////////////////////////////////////////////////

//初始化timeago组件
require(['jquery', 'jquery.timeago'], function ($) {
	$.timeago.settings.allowFuture = true;
	$.timeago.settings.localeTitle = true;
	$.timeago.settings.strings = {
		prefixAgo: null,
		prefixFromNow: null,
		suffixAgo: "前",
		suffixFromNow: "后",
		inPast: '现在',
		seconds: "<1分钟",
		minute: "1分钟",
		minutes: "%d 分钟",
		hour: "1小时",
		hours: "%d 小时",
		day: "一天",
		days: "%d 天",
		month: "一个月",
		months: "%d 个月",
		year: "一年",
		years: "%d 年",
		wordSeparator: " ",
		numbers: []
	};
});

///////////////////////////////////////////////////////////////////////////////////////////////
// 全局监听
///////////////////////////////////////////////////////////////////////////////////////////////

//导航条
require(['jquery'], function ($) {
	//一级导航条
	$('.m-nav').on('mouseenter mouseleave', '.j-drop', function (event) {
		switch (event.type) {
		case 'mouseenter':
			$(this).find(".j-drop-content").show();
			break;
		case 'mouseleave':
			$(this).find(".j-drop-content").hide();
			break;
		}
	});
	//二级拓展条
	$('.m-nav').on('mouseenter mouseleave', '.j-right-drop', function (event) {
		switch (event.type) {
		case 'mouseenter':
			$(this).find(".j-right-drop-content").show();
			break;
		case 'mouseleave':
			$(this).find(".j-right-drop-content").hide();
			break;
		}
	});
});


// 导航条处理
// 鼠标移动到drop出现下拉菜单
require(['jquery', 'jquery.timeago'], function ($) {
	var header_message = false;
	/*
		if (store.get("read") > 0 || header_message == false) {
			var newNumber = parseInt(store.get("read"));
			header_message = true;
			store.set("read", 0);
			$(".m-nav .info-number").text("").hide();
			$.ajax({
				url: "/web/admin/account_messagepost",
				type: "POST",
				data: {
					page: 1,
					clear: "true"
				},
				beforeSend: function () {
					_this.find(".j-drop-content").html("<li class='f-tac f-p20 text-muted'>消息获取中&nbsp;<i class='fa fa-refresh fa-spin'></i></li>");
				},
				success: function (data, textStatus) {
					var c = _this.find(".j-drop-content");
					c.html("");
					if (data == "") {
						c.html("<li class='f-tac f-p20 text-muted'>暂无消息</li>");
						return false;
					}
					for (var i = 0; i < data.length; i++) {
						var title = getMessageType(data[i].Message.Type, data[i].Message.Link);
						var description = "<span class='f-ml10'>" + data[i].Message.Description + "</span><span class='f-ml10 timeago' title='" + data[i].Time + "'></span>";
						var link = "<a href='" + data[i].Message.Link + "' target='_blank'>点击查看</a>";
						c.append("<a href='" + data[i].Message.Link + "' id='message-" + i + "' class='message-content f-cb'>" + title + description + "</a>");
						c.append("<li class='cut'></li>");
						//显示友好时间
						$(".timeago").timeago();
					}
					c.append("<a href='user.html?to=/web/admin/account_message' class='f-cb f-tac f-wm'>更多消息</a>");
					//让最新的消息闪一下
					for (var i = 0; i < newNumber; i++) {
						$("#message-" + i).css("font-weight", "bold");
					}
					setTimeout(function () {
						$(".m-nav .message-content").removeAttr("style");
					}, 2000);
				}
			});
		}
		break;
	}
	*/
});

//导航栏自动隐藏
var autoHidePreHeight = 0;
var autoHideFlag = false;
var forceHideNav = false;
require(['jquery'], function ($) {
	$(window).scroll(function () {
		//是否禁用
		if (forceHideNav == true) {
			$(".m-nav").hide();
			return;
		}

		$(".m-nav").show();

		if ($(window).scrollTop() <= 40) {
			if (autoHideFlag) {
				autoHideFlag = false;
				$('.m-nav').css('top', '0px');
			}
			autoHidePreHeight = $(window).scrollTop();
			return
		}
		if ($(window).scrollTop() > autoHidePreHeight && !autoHideFlag) {
			autoHideFlag = true;
			$('.m-nav').css('top', '-40px');
		} else if ($(window).scrollTop() < autoHidePreHeight && autoHideFlag) {
			autoHideFlag = false;
			$('.m-nav').css('top', '0px');
		}
		autoHidePreHeight = $(window).scrollTop();
	});
});

///////////////////////////////////////////////////////////////////////////////////////////////
// global - 全局vm
///////////////////////////////////////////////////////////////////////////////////////////////

var global = avalon.define({
	$id: "global",
	my: {}, // 我的信息
	myLogin: false, // 是否已登陆
	temp: {
		myDeferred: null // 我的信息执行状态
	}, // 缓存
	emptyObject: function (obj) { //判断对象是否为空
		require(['jquery'], function ($) {
			return $.isEmptyObject(obj);
		});
	},
	getMessage: function () { //获取用户消息
		console.log('get message');
	},
	signout: function () { //退出登陆
		post('/api/check/signout', null, '已退出', null, function (data) {
			global.myLogin = false;
			global.my = {};
			//如果用户在用户信息后台则返回首页
			if (mmState.currentState.stateName.indexOf('user') > -1) {
				avalon.router.navigate('/');
			}
		});
	},
	$skipArray: ['emptyObject', 'temp']
});

///////////////////////////////////////////////////////////////////////////////////////////////
// 状态路由
///////////////////////////////////////////////////////////////////////////////////////////////

require(['jquery', 'mmState'], function ($) {
	//获取登陆用户信息
	global.temp.myDeferred = $.Deferred();
	post('/api/currentUser', null, null, null, function (data) {
		data.image = userImage(data.image);
		global.my = data;
		global.myLogin = true;
		global.temp.myDeferred.resolve(); // 信息获取完毕 用户已登录
	}, function () {
		global.temp.myDeferred.resolve(); // 信息获取完毕 用户未登录
	});

	//找不到的页面跳转到404
	avalon.router.error(function () {
		avalon.router.navigate('/404');
	});

	//模版无法加载跳转404
	avalon.state.config({
		onloadError: function () {
			avalon.router.navigate("/404");
		},
		onBeforeUnload: function () {
			// 清空所有jbox
			$('.jBox-wrapper').remove()
		}
	})

	//404
	avalon.state("404", {
		controller: "global",
		url: "/404",
		views: {
			"container": {
				templateUrl: '/static/html/public/404.html',
				controllerUrl: ['js/public/404.js'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	});

	//首页
	avalon.state("index", {
		controller: "global",
		url: "/",
		views: {
			"container": {
				templateUrl: '/static/html/index/home.html',
				controllerUrl: ['js/index/index'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//登陆
	avalon.state("login", {
		controller: "global",
		url: "/login",
		views: {
			"container": {
				templateUrl: '/static/html/check/login.html',
				controllerUrl: ['js/check/login'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//第三方平台登陆
	avalon.state("loginOauth", {
		controller: "global",
		url: "/login/oauth",
		views: {
			"container": {
				templateUrl: '/static/html/check/loginOauth.html',
				controllerUrl: ['js/check/loginOauth'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//注册
	avalon.state("register", {
		controller: "global",
		url: "/register",
		views: {
			"container": {
				templateUrl: '/static/html/check/register.html',
				controllerUrl: ['js/check/register'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 自动验证
	avalon.state("auth", {
		controller: "global",
		url: "/auth",
		views: {
			"container": {
				templateUrl: '/static/html/check/auth.html',
				controllerUrl: ['js/check/auth'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 创建网站
	avalon.state("create", {
		controller: "global",
		url: "/create",
		views: {
			"container": {
				templateUrl: '/static/html/create/create.html',
				controllerUrl: ['js/create/create'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//游戏列表
	avalon.state("gameList", {
		controller: "global",
		url: "/games",
		views: {
			"container": {
				templateUrl: '/static/html/game/game.html',
				controllerUrl: ['js/game/game'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 板块
	avalon.state("game", {
		controller: "global",
		url: "/g/{game}",
		views: {
			"container": {
				templateUrl: '/static/html/game/base.html',
				controllerUrl: ['js/game/base'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		},
		abstract: true
	})

	//板块.首页
	avalon.state("game.home", {
		controller: "gameBase",
		url: "",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/home.html',
				controllerUrl: ['js/game/home'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 管理
	avalon.state("game.admin", {
		controller: "gameBase",
		url: "/admin",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/admin.html',
				controllerUrl: ['js/game/admin'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		},
		abstract: true
	})

	// 管理 首页
	avalon.state("game.admin.home", {
		controller: "gameAdmin",
		url: "",
		views: {
			"gameAdminContainer": {
				templateUrl: '/static/html/game/adminHome.html',
				controllerUrl: ['js/game/adminHome'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 管理 具体项目
	avalon.state("game.admin.info", {
		controller: "gameAdmin",
		url: "/{info}",
		views: {
			"gameAdminContainer": {
				templateUrl: function (param) {
					return '/static/html/game/admin/' + param.info + '.html';
				},
				controllerUrl: function (param) {
					return ['js/game/admin/' + param.info];
				},
				cacheController: false
			}
		}
	})

	// 板块.标签
	avalon.state("game.tag", {
		controller: "gameBase",
		url: "/tag",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/list.html',
				controllerUrl: ['js/game/list'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//板块.分类列表
	avalon.state("game.list", {
		controller: "gameBase",
		url: "/{category:[a-z]{1,10}}",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/list.html',
				controllerUrl: ['js/game/list'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//板块.分类列表
	avalon.state("game.list", {
		controller: "gameBase",
		url: "/{category:[a-z]{1,10}}/doc",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/listDoc.html',
				controllerUrl: ['js/game/listDoc'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//板块.文章信息
	avalon.state("game.page", {
		controller: "gameBase",
		url: "/{id:[0-9a-z]{24}}",
		views: {
			"gameContainer": {
				templateUrl: '/static/html/game/page.html',
				controllerUrl: ['js/game/page'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	//板块.文档.文章信息
	avalon.state("game.list.doc", {
		controller: "gameListDoc",
		url: "/{id:[0-9a-z]{24}}",
		views: {
			"gameListDocContainer": {
				templateUrl: '/static/html/game/pageDoc.html',
				controllerUrl: ['js/game/page'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 账号后台
	avalon.state("user", {
		controller: "global",
		url: "/user",
		views: {
			"container": {
				templateUrl: '/static/html/user/base.html',
				controllerUrl: ['js/user/base'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		},
		abstract: true
	})

	// 账号后台 - 分类 - 页面
	avalon.state("user.page", {
		controller: "userBase",
		url: "/{category}/{page}",
		views: {
			"userContainer": {
				templateUrl: function (params) {
					console.log('templateUrl')
					return '/static/html/user/' + params.category + '/' + params.page + '.html';
				},
				controllerUrl: function (params) {
					//设置当前分类和页面
					avalon.vmodels.userBase.category = params.category;
					avalon.vmodels.userBase.page = params.page;

					//改变当前标题
					for (var key in avalon.vmodels.userBase.lists.$model) {
						if (avalon.vmodels.userBase.lists[key].url == params.category) {
							for (var _key in avalon.vmodels.userBase.lists[key].childs.$model) {
								if (avalon.vmodels.userBase.lists[key].childs[_key].url == params.page) {
									avalon.vmodels.userBase.title = '<i class="f-mr5 fa ' + avalon.vmodels.userBase.lists[key].childs[_key].icon + '"></i>' + avalon.vmodels.userBase.lists[key].childs[_key].name;
									document.title = '我的账号 - ' + avalon.vmodels.userBase.lists[key].childs[_key].name + ' - 我酷游戏';
								}
							}
						}
					}

					return ['js/user/' + params.category + '/' + params.page];
				},
				cacheController: false
			}
		}
	})

	// 更新/新增第三方平台
	avalon.state("oauth", {
		controller: "global",
		url: "/oauth",
		views: {
			"container": {
				templateUrl: '/static/html/check/oauth.html',
				controllerUrl: ['js/check/oauth'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 第三方平台二跳地址
	avalon.router.get("/oauth/jump", function () {
		location.href = "https://openapi.baidu.com/social/oauth/2.0/receiver" + location.search;
	})

	// 舆情分析
	avalon.state("yuqing", {
		controller: "global",
		url: "/yuqing",
		views: {
			"container": {
				templateUrl: '/static/html/yuqing/yuqing.html',
				controllerUrl: ['js/yuqing/yuqing'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 舆情分析详细列表
	avalon.state("yuqingList", {
		controller: "global",
		url: "/yuqing/{category}",
		views: {
			"container": {
				templateUrl: '/static/html/yuqing/list.html',
				controllerUrl: ['js/yuqing/list'],
				ignoreChange: function (changeType) {
					if (changeType) return true;
				}
			}
		}
	})

	// 启动路由
	avalon.history.start({
		basepath: "/",
		html5Mode: true,
		hashPrefix: '!'
	})

	// 扫描
	avalon.scan()
});