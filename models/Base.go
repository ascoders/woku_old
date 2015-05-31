package models

import (
	"bytes"
	"encoding/gob"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"github.com/hoisie/redis"
	"gopkg.in/mgo.v2"
)

var (
	Db    *mgo.Database //数据库
	Redis redis.Client  //redis
)

type Base struct {
	Id bson.ObjectId `json:"_id" form:"_id" bson:"_id" valid:"Required"` //主键
}

func init() {
	//获取数据库连接
	session, err := mgo.Dial(beego.AppConfig.String("MongoDb"))
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	Db = session.DB("ascode")
	//初始化redis数据库
	Redis.Addr = "127.0.0.1:6379"
}

func (this *Base) Count() int {
	count, err := Db.C("card").Count()
	if err != nil {
		return 0
	}
	return count
}

/* 结构体编码为字节流 */
func StructEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/* 字节流解码为结构体 */
func StructDecode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

/* 自动读存缓存 */
func AutoCache(key string, obj interface{}, time int64, callback func()) {
	if cache, err := Redis.Get(key); err == nil { //存在缓存
		StructDecode(cache, obj)
	} else { //没有缓存
		callback()
		if obj == nil {
			return
		}
		cache, _ := StructEncode(obj)
		Redis.Setex(key, time, cache)
	}
}

/* 删除某个缓存 */
func DeleteCache(key string) {
	Redis.Del(key)
}

/* 根据前缀删除缓存 */
func DeleteCaches(prefix string) {
	keys, _ := Redis.Keys(prefix + "*")

	for k, _ := range keys {
		Redis.Del(keys[k])
	}
}

/* model的interface接口 */
type BaseInterface interface {
	BaseFormStrings(contro *beego.Controller) (bool, interface{}) //解析数组
	BaseCount() int
	BaseFilterCount(filterKey string, filter interface{}) int
	BaseInsert() error
	BaseFind(id string) (bool, interface{})
	BaseUpdate() error
	BaseDelete() error
	BaseSetId(id string) (bool, interface{})
	BaseSelect(from int, limit int, sort string) []BaseInterface
	BaseFilterSelect(from int, limit int, sort string, filterKey string, filter interface{}) []BaseInterface
	BaseSelectLike(from int, limit int, sort string, target string, key interface{}) []BaseInterface
	BaseLikeCount(target string, key interface{}) int
	BaseSelectAccuracy(from int, limit int, sort string, target string, key interface{}) []BaseInterface
	BaseAccuracyCount(target string, key interface{}) int
}

/* 后台操作数据表 */
func Restful(this BaseInterface, contro *beego.Controller) {
	ok, data := func() (bool, interface{}) {
		switch contro.GetString("type") {
		case "add": //增
			if err := contro.ParseForm(this); err != nil {
				return false, err.Error()
			}

			//数据验证
			valid := validation.Validation{}
			b, err := valid.Valid(this)
			if err != nil {
				return false, "验证参数解析失败"
			}

			if !b { //验证失败
				for _, err := range valid.Errors {
					return false, err.Key + " " + err.Message
				}
			}

			// 额外解析数组参数
			if _ok, _data := this.BaseFormStrings(contro); !_ok {
				return false, _data
			}

			if err := this.BaseInsert(); err != nil {
				return false, err.Error()
			}

			return true, "" // 插入成功
		case "delete": //删
			if _ok, _data := this.BaseSetId(contro.GetString("_id")); !_ok {
				return false, _data
			}

			err := this.BaseDelete()

			if err != nil {
				return false, err
			}

			return true, nil
		case "update": //改
			//根据id查询对象
			if _ok, _data := this.BaseFind(contro.GetString("_id")); !_ok {
				return false, _data
			}

			//解析请求参数->对象
			if err := contro.ParseForm(this); err != nil {
				return false, err.Error()
			}

			// 额外解析数组参数
			if _ok, _data := this.BaseFormStrings(contro); !_ok {
				return false, _data
			}

			//验证
			valid := validation.Validation{}
			b, err := valid.Valid(this)
			if err != nil {
				return false, "验证参数解析失败"
			}

			if !b { //验证失败
				for _, err := range valid.Errors {
					return false, err.Key + " " + err.Message
				}
			}

			if err := this.BaseUpdate(); err != nil {
				return false, err.Error()
			}
			return true, nil
		case "get": //查
			from, _ := contro.GetInt("from")
			number, _ := contro.GetInt("number")

			if number > 100 {
				return false, "最多查询100条"
			}

			if contro.GetString("like") != "" && contro.GetString("likeKey") != "" { // 模糊搜索
				var result []BaseInterface
				var count int

				if contro.GetString("likeMethod") == "like" {
					result = this.BaseSelectLike(from, number, contro.GetString("sort"), contro.GetString("likeKey"), contro.GetString("like"))
					count = this.BaseLikeCount(contro.GetString("likeKey"), contro.GetString("like"))
				} else if contro.GetString("likeMethod") == "accuracy" {
					result = this.BaseSelectAccuracy(from, number, contro.GetString("sort"), contro.GetString("likeKey"), contro.GetString("like"))
					count = this.BaseAccuracyCount(contro.GetString("likeKey"), contro.GetString("like"))
				}

				return true, map[string]interface{}{
					"lists": result,
					"count": count,
				}
			} else { // 普通查询
				if contro.GetString("filter") != "" {
					filter, _ := contro.GetInt("filter")
					count := this.BaseFilterCount(contro.GetString("filterKey"), filter)
					result := this.BaseFilterSelect(from, number, contro.GetString("sort"), contro.GetString("filterKey"), filter)

					return true, map[string]interface{}{
						"lists": result,
						"count": count,
					}
				} else {
					count := this.BaseCount()
					result := this.BaseSelect(from, number, contro.GetString("sort"))

					return true, map[string]interface{}{
						"lists": result,
						"count": count,
					}
				}
			}
		}
		return true, nil
	}()

	contro.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	contro.ServeJson()
}
