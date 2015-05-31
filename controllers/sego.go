/**
  分词
*/

package controllers

import (
	"github.com/astaxie/beego"
	"github.com/huichen/sego"
	"regexp"
	"strings"
	"woku/models"
)

type SegoController struct {
	beego.Controller
}

var segmenter sego.Segmenter

var segoLoading bool

func (this *SegoController) ToSlices(txt string, searchMode bool) []string {
	text := []byte(txt)
	segments := segmenter.Segment(text)

	// 分词
	result := sego.SegmentsToSlice(segments, searchMode)

	// 去重
	in := map[string]string{}

	for k, _ := range result {
		if _, ok := in[result[k]]; !ok { //不存在值才加入

			// 只允许中文、字母
			if matched, _ := regexp.MatchString("[a-z\u4e00-\u9fa5]", result[k]); !matched {
				continue
			}

			in[result[k]] = strings.TrimSpace(result[k])
		}
	}

	// 生成新数组
	var newResult []string

	for k, _ := range in {
		newResult = append(newResult, in[k])

	}

	return newResult
}

func (this *SegoController) ToSegments(txt string) []sego.Segment {
	text := []byte(txt)
	segments := segmenter.Segment(text)

	return segments
}

// 加载字典
func (this *SegoController) LoadDictionary() {
	if segoLoading {
		return
	}

	// 载入词典
	segoLoading = true
	// 读取全部分词
	split := &models.YuqingSplit{}
	result := split.All()

	interfaces := make([]sego.Data, len(result))
	for k, _ := range result {
		interfaces[k] = result[k]
	}

	segmenter.LoadDict(interfaces) //载入词典
	segoLoading = false
}
