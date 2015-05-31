package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Article struct {
	Id       bson.ObjectId `bson:"_id"` //主键
	Uid      bson.ObjectId `bson:"u"`   //用户id
	Category uint16        `bson:"c"`   //所属分类id
	Title    string        `bson:"t"`   //标题
	Content  string        `bson:"co"`  //正文内容
	Views    uint32        `bson:"v"`   //浏览量
	Source   string        `bson:"s"`   //来源（百度）
	Link     string        `bson:"l"`   //来源地址
	Time     time.Time     `bson:"tm"`  //发布日期
}

var (
	articleC *mgo.Collection //数据库连接
)

func init() {
	articleC = Db.C("article")
}

/* 使用缓存 */
func (this *Article) Cache() *Article {
	return this
}

/* 插入文章 */
func (this *Article) Insert() string {
	this.Id = bson.NewObjectId()
	this.Time = time.Now()
	err := articleC.Insert(this)
	if err != nil {
		return ""
	}

	return bson.ObjectId.Hex(this.Id)
}

/* 查询某个标题是否存在 */
func (this *Article) FindTitle(title string) bool {
	var ok bool
	AutoCache("article-FindTitle"+title, &ok, 60, func() {
		err := articleC.Find(bson.M{"t": title}).One(&this)
		if err == nil {
			ok = true
		}
	})
	return ok
}

/* 插入舆情分析的文章 */
func (this *Article) InsertSituation() string {
	this.Id = bson.NewObjectId()
	this.Uid = bson.ObjectIdHex("53e080d21fc9010589000002")
	err := articleC.Insert(this)
	if err != nil {
		return ""
	}

	return bson.ObjectId.Hex(this.Id)
}

/* 更新文章 */
func (this *Article) Update(change bson.M) {
	colQuerier := bson.M{"_id": this.Id}
	articleC.Update(colQuerier, change)
}

/* 根据ObjectId查询某个文章信息 */
func (this *Article) FindOne(value string) bool {
	ok := true
	AutoCache("article-FindOne"+value, &this, 60, func() {
		if !bson.IsObjectIdHex(value) { //检查id格式是否正确
			ok = false
			return
		}
		err := articleC.Find(bson.M{"_id": bson.ObjectIdHex(value)}).One(&this)
		if err != nil {
			ok = false
		}
	})
	return ok
}

/* 查询某些分类下的文章某些信息 */
func (this *Article) FindArticles(categorys []uint16, from int, to int, content bool) []*Article {
	var articles []*Article
	var err error
	if content {
		AutoCache("article-FindArticles"+fmt.Sprintln(categorys, from, to, content), &this, 60*60, func() {
			err = articleC.Find(bson.M{"c": bson.M{"$in": categorys}}).Sort("-_id").Skip(from).Limit(to - from).All(&articles)
		})
	} else {
		AutoCache("article-FindArticles"+fmt.Sprintln(categorys, from, to, content), &this, 60*60, func() {
			err = articleC.Find(bson.M{"c": bson.M{"$in": categorys}}).Select(bson.M{"co": 0}).Sort("-_id").Skip(from).Limit(to - from).All(&articles)
		})
	}
	if err != nil {
		return nil
	} else {
		return articles
	}
}

/* 查询某些分类下热门文章信息 */
func (this *Article) FindHots(categorys []uint16, limit int) []*Article {
	var articles []*Article
	AutoCache("article-FindHots"+fmt.Sprintln(categorys, limit), &articles, 60*60, func() {
		articleC.Find(bson.M{"c": bson.M{"$in": categorys}}).Sort("-v").Limit(limit).
			Select(bson.M{"_id": 1, "u": 1, "t": 1, "v": 1}).All(&articles)
	})
	return articles
}

/* 查询某些分类下文章总数 */
func (this *Article) FindCount(categorys []uint16) int {
	var count int
	AutoCache("article-FindCount"+fmt.Sprintln(categorys), &count, 60, func() {
		count, _ = articleC.Find(bson.M{"c": bson.M{"$in": categorys}}).Count()
	})
	return count
}

/* 查询某个用户前i~i*20个文章 */
func (this *Article) FindUserArticles(id bson.ObjectId, from uint32, to uint32) []*Article {
	var result []*Article
	AutoCache("article-FindUserArticles"+fmt.Sprintln(id, from, to), &result, 60, func() {
		articleC.Find(bson.M{"u": id}).Sort("-_id").Skip(int(from)).Limit(int(to - from)).
			Select(bson.M{"_id": 1, "u": 1, "t": 1, "v": 1}).All(&result)
	})
	return result
}

/* 查询某用户的文章总数 */
func (this *Article) FindUserArticleCount(id bson.ObjectId) int {
	var count int
	AutoCache("article-FindUserArticleCount"+id.Hex(), &count, 60, func() {
		count, _ = articleC.Find(bson.M{"u": id}).Count()
	})
	return count
}

/* 删除文章 */
func (this *Article) Delete() {
	articleC.Remove(bson.M{"_id": this.Id})
}

/* 查询文章总数 */
func (this *Article) Count() int {
	var count int
	AutoCache("article-Count", &count, 60, func() {
		count, _ = articleC.Count()
	})
	return count
}

/* 查询 sitemap*/
func (this *Article) FindSitemap(from int, number int) []*Article {
	var result []*Article
	AutoCache("article-FindSitemap", &result, 60, func() {
		articleC.Find(bson.M{}).Sort("_id").Skip(from).Limit(number).Select(bson.M{"_id": 1}).All(&result)
	})
	return result
}

/* 查询所有文章，包括浏览量*/
func (this *Article) FindAllAndViews() []*Article {
	var result []*Article
	AutoCache("article-FindAllAndViews", &result, 60*60, func() {
		articleC.Find(bson.M{}).Select(bson.M{"_id": 1, "v": 1}).All(&result)
	})
	return result
}

/* 为某个文章增加浏览数 */
func (this *Article) AddViews(id string) bool {
	//如果不是bsonid则退出
	if !bson.IsObjectIdHex(id) {
		return false
	}
	if err := articleC.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$inc": bson.M{"v": 1}}); err != nil {
		return false
	}
	return true
}

/* 查询某个分类的浏览总量 */
func (this *Article) FindViews(category uint16) uint64 {
	var views uint64
	AutoCache("article-FindViews"+fmt.Sprintln(category), &views, 1, func() {
		//查找该分类下全部文章
		var result []*Article
		err := articleC.Find(bson.M{"c": category}).Select(bson.M{"v": 1}).All(&result)
		fmt.Println(len(result))
		if err != nil {
			return
		}
		//统计浏览总数
		views := uint64(0)
		for k, _ := range result {
			views += uint64(result[k].Views)
		}
	})
	return views
}

/* 查询最新文章 */
func (this *Article) FindTop(limit int) []*Article {
	var result []*Article
	AutoCache("article-FindTop"+fmt.Sprintln(limit), &result, 60*60, func() {
		articleC.Find(bson.M{}).Sort("-_id").Select(bson.M{"_id": 1, "t": 1, "c": 1, "co": 1}).Limit(limit).All(&result)
	})
	return result
}

/* 查询资讯5个最新文章 */
func (this *Article) FindNews() []*Article {
	var result []*Article
	AutoCache("article-FindNews", &result, 60*60, func() {
		articleC.Find(bson.M{"c": 11}).Sort("-_id").Select(bson.M{"_id": 1, "t": 1, "c": 1, "co": 1}).Limit(5).All(&result)
	})
	return result
}
