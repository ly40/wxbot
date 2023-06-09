package word

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/pkg/log"
	"github.com/yqchilde/wxbot/engine/robot"
)

// https://www.free-api.com/doc/580
const (
	WordURL = "https://www.mxnzp.com/api/convert/dictionary?content=%s&app_id=rgihdrm0kslojqvm&app_secret=WnhrK251TWlUUThqaVFWbG5OeGQwdz09"
)

func init() {
	engine := control.Register("word", &control.Options{
		Alias: "🀄新华字典",
		Help: `指令：查字典 [字]
例如：查字典 啤`,
	})

	engine.OnRegex(`^查字典 ?(.*?)$`).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		word := ctx.State["regex_matched"].([]string)[1]
		if data, err := getWord(word); err == nil {
			if data == nil {
				ctx.ReplyTextAndAt("❌没查到该字典信息")
			} else {
				ctx.ReplyTextAndAt(fmt.Sprintf(`%s：
繁体：%s
拼音：%s
偏旁部首：%s
释义：%s
笔画：%d`, word, data.Traditional, data.Pinyin, data.Radicals, strings.ReplaceAll(data.Explanation, "\n\n", "\n"), data.Strokes))
			}
		} else {
			ctx.ReplyTextAndAt("查询失败，这一定不是bug🤔")
		}
	})
}

type wordData struct {
	Word        string `json:"word"`
	Traditional string `json:"traditional"`
	Pinyin      string `json:"pinyin"`
	Radicals    string `json:"radicals"`
	Explanation string `json:"explanation"`
	Strokes     int    `json:"strokes"`
}
type apiResponse struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []wordData `json:"data"`
}

func getWord(word string) (*wordData, error) {
	var res apiResponse

	words := url.QueryEscape(word) //做一层编码操作
	url := fmt.Sprintf(WordURL, words)
	if err := req.C().Get(url).Do().Into(&res); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, nil
	}
	var data wordData
	for _, p := range res.Data {
		if p.Word == word {
			data = p
		}
	}

	return &data, nil
}
