package controllers

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"woku/models"
)

type DocController struct {
	beego.Controller
}

// 查询某category下子文档
func (this *DocController) GetDoc() {
	ok, data := func() (bool, interface{}) {
		doc := &models.Doc{}

		if _ok := doc.Find(this.GetString("id")); !_ok {
			return false, nil
		}

		return true, doc
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 新增文档
// 返回子集id
func (this *DocController) Add(name string, isFolder bool, id string, parent string, category string, topicId string) (bool, string) {
	ok, data := func() (bool, string) {
		if strings.Trim(name, " ") == "" {
			return false, "名称不能为空"
		}

		// 实例化要添加的子文档
		docChild := models.DocChild{}
		docChild.Id = bson.ObjectIdHex(topicId)
		docChild.IsFolder = isFolder
		docChild.Name = name

		// 实例化文档
		doc := &models.Doc{}

		if _ok := doc.Find(id); _ok { // 文档存在

			// 文档子元素长度不能大于200
			if len(doc.Childs) > 200 {
				return false, "每个文件夹最多放置200个文件，超过可以拆分多个文件夹"
			}

			if doc.Nested >= 5 && isFolder {
				return false, "文件夹最多5层"
			}

			// 尾部新增
			doc.PushChild(docChild)

		} else { // 文档不存在

			// 放置新增子文档
			doc.Childs = []models.DocChild{docChild}

			// 查询其父级文档
			parentDoc := &models.Doc{}
			if _ok := parentDoc.Find(parent); !_ok { // 父文档不存在，则新增根文档
				doc.Nested = 0

				if _ok := doc.Insert(category, category, category); !_ok {
					return false, "新增文档失败"
				}
			} else { // 父文档存在，则新增文档
				doc.Nested = parentDoc.Nested + 1

				if doc.Nested >= 5 && docChild.IsFolder {
					return false, "文件夹最多5层"
				}

				// 判断此文档id是否在父文档中
				flag := false
				for k := range parentDoc.Childs {
					if parentDoc.Childs[k].Id.Hex() == id {
						flag = true
						break
					}
				}

				if !flag {
					return false, "新增文档不在父文档中"
				}

				if _ok := doc.Insert(id, category, parentDoc.Id.Hex()); !_ok {
					return false, "新增文档失败"
				}
			}

		}

		return true, topicId
	}()

	return ok, data
}

// 删除文件夹
func (this *DocController) DeleteFolder() {
	ok, data := func() (bool, interface{}) {
		// 未登录
		if this.GetSession("WOKUID") == nil {
			return false, "未登录"
		}

		// 查询用户
		member := &models.Member{}
		if ok := member.FindOne(this.GetSession("WOKUID").(string)); !ok {
			return false, "用户不存在"
		}

		// 查询父级文档
		parentDoc := &models.Doc{}
		if _ok := parentDoc.Find(this.GetString("parent")); !_ok {
			return false, "父级不存在"
		}

		// 获取分类信息
		category := &models.GameCategory{}
		if ok := category.FindId(parentDoc.Category.Hex()); !ok { // 此处parent就是category的id
			return false, "该分类不存在"
		}

		// 获取游戏信息
		game := &models.Game{}
		if ok := game.FindPath(category.Game); !ok {
			return false, "该板块不存在"
		}

		index, _ := this.GetInt("index")

		if index < 0 || index >= len(parentDoc.Childs) {
			return false, "删除位置不存在"
		}

		if !parentDoc.Childs[index].IsFolder {
			return false, "不能删除文件"
		}

		// 查询文档
		doc := &models.Doc{}
		if _ok := doc.Find(parentDoc.Childs[index].Id.Hex()); _ok { // 存在文件夹
			if len(doc.Childs) > 0 {
				return false, "不能删除非空文档"
			}
		}

		// 删除子文档
		parentDoc.DeleteChild(parentDoc.Childs[index])

		return true, nil
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 删除文件
func (this *DocController) DeleteFile(parent string, index int, topicId string) (bool, interface{}) {
	ok, data := func() (bool, interface{}) {
		// 查询父级文档
		parentDoc := &models.Doc{}
		if _ok := parentDoc.Find(parent); !_ok {
			return false, "父级不存在"
		}

		if index < 0 || index >= len(parentDoc.Childs) {
			return false, "删除位置不存在"
		}

		if parentDoc.Childs[index].IsFolder {
			return false, "不能删除一个文件夹"
		}

		if parentDoc.Childs[index].Id.Hex() != topicId {
			return false, "删除话题与文档中不一致"
		}

		// 删除子文件
		parentDoc.DeleteChild(parentDoc.Childs[index])

		return true, nil
	}()

	return ok, data
}

// 根据文章id查询之前节点信息
func (this *DocController) Parents() {
	ok, data := func() (bool, interface{}) {

		docArray := []models.Doc{}

		// 查找父级（第一次比较费时）
		doc := models.Doc{}
		if _ok := doc.IncludeChild(this.GetString("id")); !_ok {
			return false, "文档不存在"
		}

		if doc.Id != doc.Category { // 不是根文档
			docArray = append(docArray, doc)
		}

		tryNumber := 0

		// 如果有父级 递归查找父级
		var findParent func(thisDoc models.Doc)
		findParent = func(thisDoc models.Doc) {
			if tryNumber > 5 {
				return
			}
			tryNumber++

			// 查找父级
			parentDoc := models.Doc{}
			if _ok := parentDoc.Find(thisDoc.Parent.Hex()); _ok {
				if parentDoc.Id == parentDoc.Category { // 父级是根目录
					return
				}

				// 查询到了父级
				docArray = append(docArray, parentDoc)

				// 递归
				findParent(parentDoc)
			}
		}

		// 开始递归
		findParent(doc)

		return true, docArray
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}

// 更新文档排序
// id string 父文档id
// from int 起始位置
// to int 结束位置
func (this *DocController) Exchange() {
	ok, data := func() (bool, interface{}) {
		doc := &models.Doc{}

		if _ok := doc.Find(this.GetString("id")); !_ok {
			return false, "父级不存在"
		}

		from, _ := this.GetInt("from")
		to, _ := this.GetInt("to")

		if from < 0 || from >= len(doc.Childs) {
			return false, "起始位置不存在"
		}

		if to < 0 || to >= len(doc.Childs) {
			return false, "结束位置不存在"
		}

		if from == to {
			return false, "没有发生移动"
		}

		// 保存移动起始点child
		moveChild := doc.Childs[from]

		if to > from { // 往下移
			for i := from; i < to; i++ {
				doc.Childs[i] = doc.Childs[i+1]
			}
		} else { // 往上移
			for i := from; i > to; i-- {
				doc.Childs[i] = doc.Childs[i-1]
			}
		}
		doc.Childs[to] = moveChild

		// 更新子文档
		if _ok := doc.UpdateChild(doc.Childs); !_ok {
			return false, "移动失败"
		}

		return true, doc
	}()

	this.Data["json"] = map[string]interface{}{
		"ok":   ok,
		"data": data,
	}
	this.ServeJson()
}
