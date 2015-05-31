package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"math"
	"strconv"
	"strings"
)

type Category struct {
	Id           uint16 `bson:"_id" form:"id"`   //主键
	Title        string `bson:"t" form:"title"`  //分类名
	Name         string `bson:"n" form:"name"`   //分类英文名
	Parent       uint16 `bson:"p" form:"parent"` //上级分类id,空表示主分类
	ViewsHistory string `bson:"vh"`              //浏览量历史
}

var (
	categoryC *mgo.Collection //数据库连接
)

func init() {
	categoryC = Db.C("category")
}

/* 插入分类 */
func (this *Category) Insert() uint16 {
	this.Id = this.FindMaxId() + 1
	err := categoryC.Insert(this)
	if err != nil {
		panic(err)
	}

	return this.Id
}

/* 更新分类 */
func (this *Category) Update() {
	colQuerier := bson.M{"_id": this.Id}
	err := categoryC.Update(colQuerier, bson.M{"$set": bson.M{"t": this.Title, "n": this.Name}})
	if err != nil {
		//处理错误
	}
}

/* 删除分类 */
func (this *Category) Delete() {
	err := categoryC.Remove(bson.M{"_id": this.Id})
	if err != nil {
		return
	}
}

/* 根据Id查询某个分类信息 */
func (this *Category) FindOne(id uint16) bool {
	//查询分类
	if cache, err := Redis.Get("category-" + strconv.Itoa(int(id))); err == nil {
		StructDecode(cache, &this)
	} else {
		err := categoryC.Find(bson.M{"_id": id}).One(&this)
		if err != nil {
			return false
		}
		cache, _ := StructEncode(this)
		Redis.Setex("article-"+strconv.Itoa(int(id)), 60*60*24, cache)
	}
	return true
}

/* 查询id最大的字段 */
func (this *Category) FindMaxId() uint16 {
	//查询分类
	var result Category
	err := categoryC.Find(bson.M{}).Sort("-_id").One(&result)
	if err != nil {
		return 0
	}
	return result.Id
}

/* 查询所有主/字分类 */
func (this *Category) FindAll(main bool) []*Category {
	var result []*Category
	var condition bson.M
	if main {
		condition = bson.M{"p": 0}
	} else {
		condition = bson.M{"p": bson.M{"$ne": 0}}
	}
	err := categoryC.Find(condition).Sort("_id").All(&result)
	if err != nil {
		return nil
	}
	return result
}

/* 查询某个分类是否有子分类 */
func (this *Category) HasChild() bool {
	//查询子分类
	var result Category
	err := categoryC.Find(bson.M{"p": this.Id}).One(&result)
	if err != nil {
		return false
	}
	if result.Id == 0 {
		return false
	} else {
		return true
	}
}

/* 根据英文路径查找分类 from分类文章从哪开始 to分类文章结束位置 hots热门文章数量 */
func (this *Category) FindByName(name string, from int, to int, hotNumber int) (bool, *Category, []*Category, []*Article, []*Article, int) {
	//根据名称查询分类
	if cache, err := Redis.Get("category-name-" + name); err == nil {
		StructDecode(cache, &this)
	} else {
		err := categoryC.Find(bson.M{"n": name}).One(&this)
		if err != nil {
			return false, nil, nil, nil, nil, 0
		}
		cache, _ := StructEncode(this)
		Redis.Setex("category-name-"+name, 60*60*24, cache)
	}

	//查询父分类信息
	var parent *Category
	if this.Parent != 0 {
		if cache, err := Redis.Get("article-name-parent-" + name); err == nil {
			StructDecode(cache, &parent)
		} else {
			categoryC.Find(bson.M{"_id": this.Parent}).Select(bson.M{"t": 1, "n": 1}).One(&parent)
			cache, _ := StructEncode(parent)
			Redis.Setex("article-name-parent-"+name, 60*60*24, cache)
		}
	}
	//查询所有子分类信息
	var childs []*Category
	if cache, err := Redis.Get("article-name-childs-" + name); err == nil {
		StructDecode(cache, &childs)
	} else {
		err = categoryC.Find(bson.M{"p": this.Id}).All(&childs)
		cache, _ := StructEncode(childs)
		Redis.Setex("article-name-childs-"+name, 60*60, cache)
	}

	article := &Article{}
	//设置子类参数
	var params []uint16
	if cache, err := Redis.Get("article-name-params-" + name); err == nil {
		StructDecode(cache, &params)
	} else {
		if this.Parent != 0 { //如果自己是子分类，则只查找自己底下的文章
			params = make([]uint16, 1)
			params[0] = this.Id
		} else { //自己是父分类，查找所有子分类的文章
			params = make([]uint16, len(childs))
			for k, v := range childs {
				params[k] = v.Id
			}
		}
		cache, _ := StructEncode(params)
		Redis.Setex("article-name-params-"+name, 60*60, cache)
	}

	//查询from to段的文章
	var articles []*Article
	if cache, err := Redis.Get("article-name-articles-" + name + strconv.Itoa(from) + "-" + strconv.Itoa(to)); err == nil {
		StructDecode(cache, &articles)
	} else {
		articles = article.FindArticles(params, from, to, true)
		cache, _ := StructEncode(articles)
		Redis.Setex("article-name-articles-"+name+strconv.Itoa(from)+"-"+strconv.Itoa(to), 60*60, cache)
	}

	//查询热门文章
	var hots []*Article
	if cache, err := Redis.Get("article-name-hots-" + name); err == nil {
		StructDecode(cache, &hots)
	} else {
		hots = article.FindHots(params, hotNumber)
		cache, _ := StructEncode(hots)
		Redis.Setex("article-name-hots-"+name, 60*60, cache)
	}

	//查询所有文章总数
	var allNumber int
	if cache, err := Redis.Get("article-name-allNumber-" + name); err == nil {
		StructDecode(cache, &allNumber)
	} else {
		allNumber = article.FindCount(params)
		cache, _ := StructEncode(allNumber)
		Redis.Setex("article-name-allNumber-"+name, 60*60, cache)
	}

	//计算总页数
	allPage := math.Ceil(float64(allNumber) / float64(to-from))
	return true, parent, childs, articles, hots, int(allPage)
}

/* 查询文章的分类，分类的父级分类，同分类下其他文章 */
func (this *Category) FindParentAndArticles() (*Category, []*Article, []*Article) {
	//查询父分类信息
	var parent *Category
	if this.Parent != 0 {
		if cache, err := Redis.Get("article-parent-" + this.Title); err == nil {
			StructDecode(cache, &parent)
		} else {
			err := categoryC.Find(bson.M{"_id": this.Parent}).Select(bson.M{"t": 1, "n": 1}).One(&parent)
			if err != nil {
				return nil, nil, nil
			}
			cache, _ := StructEncode(parent)
			Redis.Setex("article-parent-"+this.Title, 60*60*24, cache)
		}
	}

	article := &Article{}
	params := make([]uint16, 1)
	params[0] = this.Id
	//查询自己以及子分类的全部文章
	var articles []*Article
	if cache, err := Redis.Get("article-articles-" + this.Title); err == nil {
		StructDecode(cache, &articles)
	} else {
		articles = article.FindArticles(params, 0, 7, false)
		cache, _ := StructEncode(articles)
		Redis.Setex("article-articles-"+this.Title, 60*60, cache)
	}

	//查询热门文章
	var hots []*Article
	if cache, err := Redis.Get("article-hots-" + this.Title); err == nil {
		StructDecode(cache, &hots)
	} else {
		hots = article.FindHots(params, 7)
		cache, _ := StructEncode(hots)
		Redis.Setex("article-hots-"+this.Title, 60*60, cache)
	}

	return parent, articles, hots
}

/* 查询所有分类(sitemap) */
func (this *Category) FindSitemap() []*Category {
	var result []*Category
	err := categoryC.Find(bson.M{}).Sort("-_id").Select(bson.M{"n": 1}).All(&result)
	if err != nil {
		return nil
	} else {
		return result
	}
}

/* 追加总浏览量历史 */
func (this *Category) AddViewHistory(views uint64) {
	if this.ViewsHistory == "" {
		this.ViewsHistory = strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.Itoa(int(views))
	} else {
		this.ViewsHistory += ";" + strconv.Itoa(int(bson.Now().Unix())) + ":" + strconv.Itoa(int(views))
	}
	historyArray := strings.Split(this.ViewsHistory, ";")
	if len(historyArray) > 30 { //如果历史记录超过30条
		//只取前30条
		subArray := historyArray[len(historyArray)-30 : len(historyArray)]
		this.ViewsHistory = strings.Join(subArray, ";")
	}
	err := categoryC.Update(bson.M{"_id": this.Id}, bson.M{"$set": bson.M{"vh": this.ViewsHistory}})
	if err != nil {
		return
	}
}
