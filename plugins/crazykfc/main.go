package crazykfc

import (
	"math/rand"
	"time"

	"github.com/imroc/req/v3"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/pkg/log"
	"github.com/yqchilde/wxbot/engine/robot"
)

var sentence []string

func init() {
	go getCrazyKFCSentence()
	rand.Seed(time.Now().UnixNano())
	engine := control.Register("kfccrazy", &control.Options{
		Alias: "🥠kfc骚话",
		Help: `描述：奇怪的网友编了一些奇怪的骚话，让我们一起看看吧
指令：kfc骚话 -> 获取肯德基疯狂星期四骚话`,
	})

	engine.OnFullMatch("kfc骚话").SetBlock(true).Handle(func(ctx *robot.Ctx) {
		if len(sentence) > 0 {
			idx := rand.Intn(len(sentence) - 1)
			ctx.ReplyText(sentence[idx])
			sentence = append(sentence[:idx], sentence[idx+1:]...)
		} else {
			getCrazyKFCSentence()
			ctx.ReplyText("数据未加载完毕，请稍后再试")
		}
	})
}

type apiResponse struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

func getCrazyKFCSentence() {
	var data []apiResponse
	api := "https://raw.fastgit.org/Nthily/KFC-Crazy-Thursday/main/kfc.json"
	if err := req.C().Get(api).Do().Into(&data); err != nil {
		log.Errorf("kfc骚话获取失败: %v", err)
		return
	}
	sentence = make([]string, 0)
	for i := range data {
		sentence = append(sentence, data[i].Text)
	}
	return
}
