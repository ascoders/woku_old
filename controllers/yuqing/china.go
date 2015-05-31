/**
中国各省rss区分
*/

package yuqing

import (
	"github.com/SlyMarbo/rss"
	"github.com/astaxie/beego"
	"strings"
	"time"
	"woku/controllers"
	"woku/models"
)

type China struct {
	beego.Controller
}

var (
	chinaRss = "http://rss.sina.com.cn/news/china/focus15.xml"

	citys = map[string][]string{
		"北京":  []string{"北京"},
		"天津":  []string{"天津"},
		"上海":  []string{"上海"},
		"重庆":  []string{"重庆"},
		"河北":  []string{"河北", "石家庄", "辛集", "藁城", "晋州", "新乐", "鹿泉", "唐山", "遵化", "迁安", "秦皇岛", "邯郸", "武安", "邢台", "南宫", "沙河", "保定", "涿州", "定州", "安国", "高碑店", "张家口", "承德", "沧州", "泊头", "任丘", "黄骅", "河间", "廊坊", "霸州", "三河", "衡水", "冀州", "深州"},
		"河南":  []string{"河南", "辖郑州", "开封", "洛阳", "平顶山", "安阳", "鹤壁", "新乡", "焦作", "濮阳", "许昌", "漯河", "三门峡", "南阳", "商丘", "信阳", "周口", "驻马店"},
		"云南":  []string{"云南", "昆明", "曲靖", "昭通", "玉溪", "楚雄州", "红河州", "文山州", "普洱", "版纳州", "大理州", "保山", "德宏州", "丽江", "怒江州", "迪庆州", "临沧"},
		"辽宁":  []string{"辽宁", "沈阳", "大连", "鞍山", "抚顺", "本溪", "丹东", "锦州", "营口", "阜新", "辽阳", "盘锦", "铁岭", "朝阳", "葫芦岛"},
		"黑龙江": []string{"黑龙江", "哈尔滨", "齐齐哈尔", "鹤岗", "双鸭山", "鸡西", "大庆", "伊春", "牡丹江", "佳木斯市", "七台河", "黑河", "绥化", "大兴安岭"},
		"湖南":  []string{"湖南", "长沙", "岳阳", "株洲", "湘潭", "衡阳", "郴州", "永州", "邵阳", "娄底", "益阳", "常德", "张家界", "怀化", "吉首"},
		"安徽":  []string{"安徽", "合肥", "蚌埠", "芜湖", "淮南", "马鞍山", "淮北", "铜陵", "安庆", "黄山", "阜阳", "宿州", "滁州", "六安", "宣城", "巢湖", "池州", "亳州"},
		"山东":  []string{"山东", "济南", "泰安", "潍坊", "德州", "滨州", "莱芜", "青岛", "烟台", "日照", "东营", "济宁", "荷泽", "聊城", "临沂", "枣庄", "淄博", "威海"},
		"新疆":  []string{"新疆", "乌鲁木", "克拉玛依", "石河子", "阿拉尔", "图木舒克", "五家渠", "北屯", "哈密", "吐鲁番", "阿克苏", "喀什", "和田", "伊宁", "塔城", "阿勒泰", "奎屯", "博乐", "昌吉", "阜康", "库尔勒", "阿图什", "乌苏"},
		"江苏":  []string{"江苏", "南京", "徐州", "连云港", "淮安", "宿迁市", "盐城", "扬州", "泰州", "南通", "镇江", "常州", "无锡", "苏州"},
		"浙江":  []string{"浙江", "杭州市", "嘉兴市", "湖州市", "绍兴市", "宁波市", "台州市", "温州市", "金华市", "衢州市", "丽水市", "舟山市"},
		"江西":  []string{"江西", "南昌", "九江", "赣州", "萍乡", "新余", "贵溪", "丰城", "景德镇", "上饶", "樟树", "宜春", "吉安", "抚州", "高安"},
		"湖北":  []string{"湖北", "武汉", "宜昌", "黄石", "十堰", "荆州", "襄樊", "鄂州", "荆门", "孝感", "黄冈", "咸宁", "随州", "仙桃", "天门", "潜江"},
		"广西":  []string{"广西", "南宁", "柳州", "桂林", "梧州", "北海", "防城港", "钦州", "贵港", "玉林", "百色", "贺州", "河池", "来宾", "崇左"},
		"甘肃":  []string{"甘肃", "兰州", "嘉峪关", "金昌", "白银", "天水", "武威", "张掖", "酒泉", "平凉", "庆阳", "定西", "陇南"},
		"山西":  []string{"山西", "太原", "大同", "阳泉", "长治", "晋城", "朔州", "晋中", "运城", "忻州", "临汾", "吕梁"},
		"内蒙古": []string{"内蒙", "呼和浩特", "包头", "乌海", "集宁", "通辽", "赤峰", "东胜", "临河", "锡林浩特", "海拉尔", "乌兰浩特", "阿拉善左旗"},
		"陕西":  []string{"陕西", "西安", "铜川", "宝鸡", "咸阳", "渭南", "汉中", "商洛", "安康", "延安", "榆林"},
		"吉林":  []string{"吉林", "长春", "吉林", "四平", "白城", "延吉", "图们", "通化", "松原", "公主岭", "辽源", "榆树", "洮南", "龙井", "和龙", "珲春", "集安", "白山"},
		"福建":  []string{"福建", "福州", "厦门", "泉州", "漳州", "三明", "龙岩", "南平", "莆田", "宁德"},
		"贵州":  []string{"贵州", "贵阳", "安顺", "遵义", "六盘水", "毕节", "兴义", "凯里", "都匀", "铜仁", "仁怀", "赤水", "清镇", "福泉"},
		"广东":  []string{"广东", "广州", "深圳", "珠海", "汕头", "韶关", "佛山", "江门", "湛江", "茂名", "肇庆", "梅洲", "汕尾", "河源", "阳江", "清远", "东莞", "中山", "潮州", "揭阳", "云浮"},
		"青海":  []string{"青海", "西宁", "德令哈", "格尔木"},
		"西藏":  []string{"西藏", "拉萨", "日喀则", "樟木镇", "江孜", "泽当", "林芝", "那曲"},
		"四川":  []string{"四川", "成都", "绵阳", "德阳", "攀枝花", "遂宁", "南充", "广元", "乐山", "宜宾", "泸州", "达州", "广安", "巴中", "雅安", "内江", "自贡", "资阳", "眉山", "江油", "阆中", "华蓥", "万源", "崇州", "简阳", "西昌", "什邡", "彭州", "峨眉山", "都江堰", "邛崃", "广汉"},
		"宁夏":  []string{"宁夏", "石嘴山", "银川", "吴忠", "中卫", "固原"},
		"海南":  []string{"海南", "海口", "文昌", "三亚", "五指山", "琼海", "儋州", "万宁"},
		"台湾":  []string{"台湾", "台北", "台中", "台南", "高雄"},
		"香港":  []string{"香港"},
		"澳门":  []string{"澳门"},
	}
)

func init() {
	go chinaPlanTask()
}

// 定期执行计划任务
func chinaPlanTask() {
	hour := time.NewTicker(1 * time.Hour)
	day := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-hour.C:
			go func() {
				GetChinaNew()
			}()
		case <-day.C:
			go func() {
				// 聚合图表
				yuqing := &controllers.YuqingController{}
				for k, _ := range citys {
					yuqing.SetTable(k)
				}
			}()
		}
	}
}

// 获取最新动态
func GetChinaNew() {
	// 抓取rss
	feed, err := rss.Fetch(chinaRss)
	if err != nil {
		return
	}

	err = feed.Update()
	if err != nil {
		return
	}

	// 循环每一项rss数据
FindInItem:
	for k, _ := range feed.Items {
		sego := controllers.SegoController{}
		sentence := strings.TrimSpace(feed.Items[k].Title)
		content := strings.TrimSpace(feed.Items[k].Content)
		sentenceWords := sego.ToSlices(sentence, false)
		contentWords := sego.ToSlices(content, false)

		allWords := append(sentenceWords, contentWords...)

		// 循环每个分词
		for k, _ := range allWords {
			// 查找城市
			for m, _ := range citys {
				// 省份、城市名逐个搜索
				for p, _ := range citys[m] {
					if allWords[k] == citys[m][p] {
						// 情感分析
						good, details, wordSlice := controllers.Scoring(sentence)

						result := &models.YuqingResult{}

						// 插入到结果表
						result.Id = sentence
						result.Category = m
						result.Good = good
						result.Detail = details
						result.WordSlice = wordSlice
						result.Insert()

						continue FindInItem
					}
				}
			}
		}
	}
}

// 获取中国当日舆情
func (this *China) Read() {
	ok, data := func() (bool, interface{}) {
		table := &models.YuqingTable{}

		results := make(map[string]*models.YuqingTable)

		for k, _ := range citys {
			results[k] = table.FindNew(k)
		}

		return true, results
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}
