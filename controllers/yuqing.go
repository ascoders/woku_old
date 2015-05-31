package controllers

import (
	"woku/models"

	"github.com/SlyMarbo/rss"
	"github.com/astaxie/beego"
	"strings"
	"time"
)

type YuqingController struct {
	beego.Controller
}

// 科技
var rssTech = []string{
	"http://rss.sina.com.cn/tech/rollnews.xml",
	"http://rss.sina.com.cn/news/allnews/tech.xml",
	"http://rss.sina.com.cn/tech/internet/home28.xml",
	"http://rss.sina.com.cn/tech/mobile/mobile_6.xml",
	"http://rss.sina.com.cn/tech/4G.xml",
	"http://rss.sina.com.cn/tech/notebook/193_1.xml",
	"http://rss.sina.com.cn/tech/tele/gn37.xml",
	"http://rss.sina.com.cn/tech/down/down20.xml",
	"http://rss.sina.com.cn/tech/discovery/discovery.xml",
	"http://rss.sina.com.cn/tech/number/new_camera.xml",
	"http://rss.sina.com.cn/tech/elec/buy_elec.xml",
}

// 财经
var rssFinance = []string{
	"http://rss.sina.com.cn/roll/finance/hot_roll.xml",
	"http://rss.sina.com.cn/news/allnews/finance.xml",
	"http://rss.sina.com.cn/finance/jsy.xml",
	"http://rss.sina.com.cn/roll/stock/hot_roll.xml",
	"http://rss.sina.com.cn/finance/fund.xml",
	"http://rss.sina.com.cn/finance/financial.xml",
	"http://rss.sina.com.cn/finance/usstock.xml",
	"http://rss.sina.com.cn/finance/hkstock.xml",
	"http://rss.sina.com.cn/finance/future.xml",
}

// 军事
var rssMilitary = []string{
	"http://rss.sina.com.cn/roll/mil/hot_roll.xml",
	"http://rss.sina.com.cn/jczs/focus.xml",
	"http://rss.sina.com.cn/jczs/taiwan20.xml",
	"http://rss.sina.com.cn/jczs/china15.xml",
}

// 体育
var rssSport = []string{
	"http://rss.sina.com.cn/roll/sports/hot_roll.xml",
	"http://rss.sina.com.cn/news/allnews/sports.xml",
	"http://rss.sina.com.cn/sports/global/focus.xml",
	"http://rss.sina.com.cn/sports/china/focus.xml",
	"http://rss.sina.com.cn/sports/basketball/cba.xml",
}

// 娱乐
var rssEntertainment = []string{
	"http://rss.sina.com.cn/ent/hot_roll.xml",
	"http://rss.sina.com.cn/news/allnews/ent.xml",
	"http://rss.sina.com.cn/ent/music/focus12.xml",
	"http://rss.sina.com.cn/ent/star/focus7.xml",
}

// 文化教育
var rssEducation = []string{
	"http://rss.sina.com.cn/roll/edu/hot_roll.xml",
	"http://rss.sina.com.cn/edu/focus19.xml",
	"http://rss.sina.com.cn/edu/exam.xml",
}

var (
	anaylseMap map[string]struct {
		Type   []int
		Ignore []int
		Value  int
	}

	anaylseLoading bool
	rssLoading     bool
	resultLoading  bool

	// 聚合各分类
	yuqingcategoryLists = []string{
		"all",
		"tech",
		"finance",
		"military",
		"sport",
		"entertainment",
		"education",
	}
)

func init() {
	yuqing := YuqingController{}

	yuqing.LoadSegoLocal()
	yuqing.LoadAnaylseLocal()

	go YuqingPlanTask()
}

/* 定期执行计划任务 */
func YuqingPlanTask() {
	timer := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-timer.C: //每6小时定时任务
			go func() {
				yuqing := YuqingController{}

				// 抓取rss
				yuqing.GetRssLocal()

				// 刷新一遍舆情图表
				for k, _ := range yuqingcategoryLists {
					yuqing.SetTable(yuqingcategoryLists[k])
				}
			}()
		}
	}
}

// RSS抓取
func FetchRss(url string, category string) {
	feed, err := rss.Fetch(url)
	if err != nil {
		return
	}

	err = feed.Update()
	if err != nil {
		return
	}

	for k, _ := range feed.Items {
		words := strings.TrimSpace(feed.Items[k].Title)

		good, details, wordSlice := Scoring(words)

		result := &models.YuqingResult{}

		// 插入到结果表
		result.Id = words
		result.Category = category
		result.Good = good
		result.Detail = details
		result.WordSlice = wordSlice
		result.Insert()
	}
}

// 中文分词Restful操作
func (this *YuqingController) Split() {
	//查询用户
	member := &models.Member{}

	var session interface{}
	if session = this.GetSession("WOKUID"); session == nil {
		return
	}

	if ok := member.FindOne(session.(string)); !ok {
		return
	}

	if !inArray("yuqing", member.Power) {
		return
	}

	// restful操作接口
	yuqingSplit := &models.YuqingSplit{}
	models.Restful(yuqingSplit, &this.Controller)
}

// 语义分析Restful操作
func (this *YuqingController) Analyse() {
	//查询用户
	member := &models.Member{}

	var session interface{}
	if session = this.GetSession("WOKUID"); session == nil {
		return
	}

	if ok := member.FindOne(session.(string)); !ok {
		return
	}

	if !inArray("yuqing", member.Power) {
		return
	}

	// restful操作接口
	yuqingAnalyse := &models.YuqingAnalyse{}
	models.Restful(yuqingAnalyse, &this.Controller)
}

// 提供各分类信息接口
func (this *YuqingController) Index() {
	ok, data := func() (bool, interface{}) {
		from, _ := this.GetInt("from")
		number, _ := this.GetInt("number")

		if number == 0 {
			number = 20
		}

		result := &models.YuqingResult{}
		lists := result.Find(this.Ctx.Input.Param(":category"), from, number)

		return true, lists
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 载入分词词库
// 内部函数
func (this *YuqingController) LoadSegoLocal() {
	sego := SegoController{}
	sego.LoadDictionary()
}

// 载入分词词库
func (this *YuqingController) LoadSego() {
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

		if !inArray("yuqing", member.Power) {
			return false, "没有舆情分析权限"
		}

		go this.LoadSegoLocal()

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 载入分析词库
// 内部函数
func (this *YuqingController) LoadAnaylseLocal() {
	if anaylseLoading {
		return
	}

	anaylseLoading = true

	// 获取全部分析词语
	anaylse := &models.YuqingAnalyse{}
	anaylses := anaylse.All()

	anaylseMap = make(map[string]struct {
		Type   []int
		Ignore []int
		Value  int
	})

	for k, _ := range anaylses {
		temp := anaylseMap[anaylses[k].Id]
		temp.Type = anaylses[k].Type
		temp.Value = anaylses[k].Power
		temp.Ignore = anaylses[k].IgnoreType
		anaylseMap[anaylses[k].Id] = temp
	}

	anaylseLoading = false
}

// 载入分析词库
func (this *YuqingController) LoadAnaylse() {
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

		if !inArray("yuqing", member.Power) {
			return false, "没有舆情分析权限"
		}

		this.LoadAnaylseLocal()

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 抓取Rss
// 内部函数
func (this *YuqingController) GetRssLocal() {
	if rssLoading {
		return
	}

	rssLoading = true

	for _, v := range rssTech {
		FetchRss(v, "tech")
	}

	for _, v := range rssFinance {
		FetchRss(v, "finance")
	}

	for _, v := range rssMilitary {
		FetchRss(v, "military")
	}

	for _, v := range rssSport {
		FetchRss(v, "sport")
	}

	for _, v := range rssEntertainment {
		FetchRss(v, "entertainment")
	}

	for _, v := range rssEducation {
		FetchRss(v, "education")
	}

	// 测试数据
	datas := []string{
		"没有人不被“海鸥老人”的故事感动",
		"韩媒：中国若不能遏制朝鲜就没理由反对韩国反导",
		"不，我还没有输",
		"小王虽然落魄，但任然是个孝子",
		"暗黑破坏神3终于通过了审批",
		"这篇草案没有得到通过",
		"学生病了可以通过老师请假",
		"学生病了可以请假，通过老师",
		"你的头脑好好啊",
		"虽然演的好烂，但长得还挺好",
		"这件事真的不行，听到了不？",
		"保障基础设施很重要,大部分东亚国家医疗无保障",
		"内向是一种病,学生病了要立刻通告老师",
	}

	result := &models.YuqingResult{}

	// 清除数据
	result.RemoveOld()

	for k, _ := range datas {
		words := strings.TrimSpace(datas[k])

		good, details, wordSlice := Scoring(words)

		// 插入到结果表
		result.Id = words
		result.Category = "test"
		result.Good = good
		result.Detail = details
		result.WordSlice = wordSlice
		result.Insert()
	}

	rssLoading = false
}

// 聚合分类
func (this *YuqingController) SetTable(category string) {
	result := &models.YuqingResult{}

	var results []*models.YuqingResult

	if category == "all" {
		results = result.FindToday()
	} else {
		results = result.FindCategoryToday(category)
	}

	yuqingTable := &models.YuqingTable{}

	yuqingTable.Category = category

	// 清除一星期之外数据
	yuqingTable.RemoveOld()

	for k, _ := range results {
		if results[k].Good == 0 {
			yuqingTable.Normal++
		} else if results[k].Good > 0 {
			yuqingTable.Good++
		} else {
			yuqingTable.Bad++
		}
	}

	yuqingTable.Insert()
}

// 抓取Rss
func (this *YuqingController) GetRss() {
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

		if !inArray("yuqing", member.Power) {
			return false, "没有舆情分析权限"
		}

		this.GetRssLocal()

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 形容词好坏度
func Scoring(words string) (int, []models.YuqingResultWord, []string) {
	segoContro := SegoController{}

	details := segoContro.ToSegments(words)

	good := 0

	// 否定极性
	notPolarity := 1

	// 确定的形容词所在位置
	sureAdjective := -1
	// 确定的名词所在位置
	sureNoun := -1

	// 返回分词词组
	wordSlice := make([]string, len(details))

	var results []models.YuqingResultWord

	var setResults = func(index int, val int, _type int) {
		detail := models.YuqingResultWord{
			Content:  details[index].Token().Text(),
			Good:     val,
			Type:     _type,
			Position: index,
		}
		results = append(results, detail)
	}

	//遇到名词，且自己不能作动词，作形容词

	// 是否为xx词性的词
	var isAnaylse = func(index int, _type string) bool {
		switch _type {
		case "断句":
			if details[index].Token().Pos() == "x" &&
				details[index].Token().Text() != "\"" &&
				details[index].Token().Text() != "“" &&
				details[index].Token().Text() != "”" {
				return true
			}
		case "形容词":
			if val, ok := anaylseMap[details[index].Token().Text()]; (ok &&
				inIntArray(0, val.Type)) || details[index].Token().Pos() == "a" {
				return true
			}
		case "名词":
			if val, ok := anaylseMap[details[index].Token().Text()]; (ok &&
				inIntArray(1, val.Type)) || details[index].Token().Pos() == "n" {
				return true
			}
		case "动词":
			if val, ok := anaylseMap[details[index].Token().Text()]; (ok &&
				inIntArray(2, val.Type)) || details[index].Token().Pos() == "v" {
				return true
			}
		case "副词":
			if val, ok := anaylseMap[details[index].Token().Text()]; ok &&
				inIntArray(3, val.Type) {
				return true
			}
		case "否定词":
			if val, ok := anaylseMap[details[index].Token().Text()]; ok &&
				inIntArray(4, val.Type) {
				return true
			}
		case "介词":
			if val, ok := anaylseMap[details[index].Token().Text()]; ok &&
				inIntArray(5, val.Type) {
				return true
			}
		case "助词":
			if val, ok := anaylseMap[details[index].Token().Text()]; ok &&
				inIntArray(6, val.Type) {
				return true
			}
		}
		return false
	}

	// 设置为xx性质的词
	var setAnaylse = func(index int, _type string) {
		if _type == "断句" {
			setResults(index, 0, -1)
			// 恢复否定极性
			notPolarity = 1

			return
		}

		if val, ok := anaylseMap[details[index].Token().Text()]; ok {
			_good := val.Value

			switch _type {
			case "形容词":
				// 是否无视情感
				if inIntArray(0, val.Ignore) {
					_good = 0
				}

				setResults(index, _good, 0)

				// 计算情感 恢复否定极性
				good += _good * notPolarity
				notPolarity = 1
			case "名词":
				// 是否无视情感
				if inIntArray(1, val.Ignore) {
					_good = 0
				}

				setResults(index, _good, 1)

				// 若该名词有情感倾向
				if _good != 0 {
					// 恢复否定极性 计算情感
					good += _good * notPolarity
					notPolarity = 1
				}
			case "动词":
				// 是否无视情感
				if inIntArray(2, val.Ignore) {
					_good = 0
				}

				setResults(index, _good, 2)

				// 若该动词有情感倾向，
				if _good != 0 {
					// 恢复否定极性 计算情感
					good += _good * notPolarity
					notPolarity = 1
				}
			case "副词":
				setResults(index, _good, 3)
			case "否定词":
				setResults(index, _good, 4)
				// 乘以极性
				notPolarity *= val.Value
			case "介词":
				setResults(index, _good, 5)
			case "助词":
				setResults(index, _good, 6)
			}
		}
	}

	// 往前搜索
	var searchBefore = func(index int, noWord func() bool, callback func(key int) bool) bool {
		if index == 0 {
			if noWord() {
				return true
			}
		} else {
			if index-1 >= 0 {
				for m := index - 1; m >= 0; m-- {
					if callback(m) {
						return true
					}
				}
			}
		}

		return false
	}

	// 往后搜索
	var searchAfter = func(index int, noWord func() bool, callback func(key int) bool) bool {
		n := len(details)

		if index == n-1 {
			if noWord() {
				return true
			}
		} else {
			for m := index + 1; m < n; m++ {
				if callback(m) {
					return true
				}
			}
		}

		return false
	}

	// 逐个分词解析
SegoLoop:
	for k, _ := range details {
		wordSlice[k] = details[k].Token().Text()

		if isAnaylse(k, "断句") {
			setAnaylse(k, "断句")
			continue SegoLoop
		}

		// 确定的形容词
		if k == sureAdjective {
			setAnaylse(k, "形容词")
			continue SegoLoop
		}
		// 确定的名词
		if k == sureNoun {
			setAnaylse(k, "名词")
			continue SegoLoop
		}

		if isAnaylse(k, "形容词") && isAnaylse(k, "副词") {
			if searchAfter(k, func() bool {
				setAnaylse(k, "形容词")
				return true
			}, func(m int) bool {
				if isAnaylse(m, "断句") {
					setAnaylse(k, "形容词")
					return true
				}

				if isAnaylse(m, "形容词") {
					setAnaylse(k, "副词")

					if isAnaylse(m, "副词") {
						sureAdjective = m
					}
					return true
				}
				return false
			}) {
				continue SegoLoop
			}
		}

		if isAnaylse(k, "介词") && isAnaylse(k, "动词") {
			// 后面没有分词
			if k == len(details)-1 {
				setAnaylse(k, "动词")
				continue SegoLoop
			}

			// 往前搜寻
			if k-2 >= 0 && isAnaylse(k-1, "断句") {
				for p := k - 2; p >= 0; p-- {
					if isAnaylse(p, "名词") || isAnaylse(p, "形容词") {
						break
					}

					if isAnaylse(p, "动词") {
						setAnaylse(k, "介词")
						continue SegoLoop
					}
				}
			}

			// 往后搜寻
			n := len(details)
			for m := k + 1; m < n; m++ {
				if isAnaylse(m, "断句") {
					setAnaylse(k, "动词")
					continue SegoLoop
				}

				// x+名词 => 无法确定
				if isAnaylse(m, "名词") {
					// 确定后面为名词
					sureNoun = m

					// 后面没有分词了
					if m == len(details)-1 {
						setAnaylse(k, "动词")
						continue SegoLoop
					}

					// x+名词+动词 => x=介词
					// 往后搜寻
					for p := m + 1; p < n; p++ {
						if isAnaylse(p, "断句") {
							setAnaylse(k, "动词")
							continue SegoLoop
						}

						if isAnaylse(p, "动词") {
							setAnaylse(k, "介词")
							continue SegoLoop
						}
					}
				}
			}

		}

		if isAnaylse(k, "否定词") && isAnaylse(k, "助词") {
			// 后面没有分词了，或后面是断句
			if k == len(details)-1 || isAnaylse(k+1, "断句") {
				if searchBefore(k, func() bool {
					setAnaylse(k, "否定词")
					return true
				}, func(m int) bool {
					if isAnaylse(m, "断句") {
						setAnaylse(k, "否定词")
						return true
					}

					if isAnaylse(m, "形容词") || isAnaylse(m, "动词") {
						setAnaylse(k, "助词")
						return true
					}
					return false
				}) {
					continue SegoLoop
				}
			}

			// 往后搜寻
			if searchAfter(k, func() bool {
				setAnaylse(k, "助词")
				return true
			}, func(m int) bool {
				if isAnaylse(m, "形容词") || isAnaylse(m, "动词") {
					setAnaylse(k, "否定词")
					return true
				}
				return false
			}) {
				continue SegoLoop
			}
		}

		// 动词 + 名词
		if isAnaylse(k, "动词") && isAnaylse(k, "名词") {
			// 后面没有分词了
			if k == len(details)-1 {
				setAnaylse(k, "名词")
				continue SegoLoop
			}

			// 往前搜寻
			if k-1 >= 0 {
				for m := k - 1; m >= 0; m-- {
					if isAnaylse(m, "断句") {
						setAnaylse(k, "名词")
						continue SegoLoop
					}

					if isAnaylse(m, "否定词") || isAnaylse(m, "动词") {
						setAnaylse(k, "动词")
						continue SegoLoop
					}
				}
			}

			// 往后搜寻
			n := len(details)
			for m := k + 1; m < n; m++ {
				if isAnaylse(m, "断句") {
					setAnaylse(k, "名词")
					continue SegoLoop
				}

				if isAnaylse(m, "名词") || isAnaylse(m, "副词") {
					setAnaylse(k, "动词")
					continue SegoLoop
				}
			}
		}

		if isAnaylse(k, "名词") && isAnaylse(k, "形容词") {
			// 后面没有分词了
			if k == len(details)-1 {
				setAnaylse(k, "名词")
				continue SegoLoop
			}

			// 后面是断句
			if k+1 < len(details)-1 && isAnaylse(k+1, "断句") {
				setAnaylse(k, "名词")
				continue SegoLoop
			}

			// 往前搜寻
			if k-1 >= 0 {
				for m := k - 1; m >= 0; m-- {
					if isAnaylse(m, "断句") {
						setAnaylse(k, "形容词")
						continue SegoLoop
					}

					if isAnaylse(m, "名词") && !isAnaylse(k, "动词") {
						setAnaylse(k, "形容词")
						continue SegoLoop
					}
				}
			}
		}

		// 简单解析兜底

		if isAnaylse(k, "否定词") {
			setAnaylse(k, "否定词")
			continue SegoLoop
		}

		if isAnaylse(k, "名词") {
			setAnaylse(k, "名词")
			continue SegoLoop
		}

		if isAnaylse(k, "动词") {
			setAnaylse(k, "动词")
			continue SegoLoop
		}

		if isAnaylse(k, "副词") {
			setAnaylse(k, "副词")
			continue SegoLoop
		}

		if isAnaylse(k, "形容词") {
			setAnaylse(k, "形容词")
			continue SegoLoop
		}
	}

	return good, results, wordSlice
}

// 更新结果
func (this *YuqingController) FreshResult() {
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

		if !inArray("yuqing", member.Power) {
			return false, "没有舆情分析权限"
		}

		if resultLoading {
			return false, "正在更新中"
		}
		resultLoading = true

		// 获取最近一天的数据
		result := &models.YuqingResult{}
		lists := result.FindToday()

		for k, _ := range lists {
			// 形容词
			words := lists[k].Id
			good, details, wordSlice := Scoring(words)

			// 更新
			lists[k].Detail = details
			lists[k].Good = good
			lists[k].WordSlice = wordSlice
			lists[k].Update()
		}

		resultLoading = false

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 抓取Rss
func (this *YuqingController) OperateStatus() {
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

		if !inArray("yuqing", member.Power) {
			return false, "没有舆情分析权限"
		}

		return true, map[string]bool{
			"sego":    segoLoading,
			"anaylse": anaylseLoading,
			"rss":     rssLoading,
			"result":  resultLoading,
		}
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 查询图表信息
func (this *YuqingController) Charts() {
	ok, data := func() (bool, interface{}) {
		table := &models.YuqingTable{}

		results := make(map[string][]*models.YuqingTable)

		for k, _ := range yuqingcategoryLists {
			results[yuqingcategoryLists[k]] = table.Find(yuqingcategoryLists[k])
		}

		return true, results
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}
