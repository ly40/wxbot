package oilprice

import (
	"fmt"

	"github.com/imroc/req/v3"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/robot"
)

// token从这个网址获取https://www.alapi.cn/api/view/113
// 和天气接口一个网址
const (
	OilPriceURL = "https://v2.alapi.cn/api/oil?token=eNDasu4VX4w6buwx"
)

func init() {
	engine := control.Register("oilprice", &control.Options{
		Alias: "🐳油价查询",
		Help: `指令：油价 [省份]
例如：油价 福建`,
	})

	engine.OnRegex(`^油价 ?(.*?)$`).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		province := ctx.State["regex_matched"].([]string)[1]
		if data, err := getOilPrice(province); err == nil {
			if data == nil {
				ctx.ReplyTextAndAt("❌没查到该省份油价信息")
			} else {
				ctx.ReplyTextAndAt(fmt.Sprintf(`🚭%s当前油价：
89号：￥%.2f
92号：￥%.2f
95号：￥%.2f
98号：￥%.2f
0号：￥%.2f`, province, data.O89, data.O92, data.O95, data.O98, data.O0))
			}
		} else {
			ctx.ReplyTextAndAt("查询失败，这一定不是bug🤔")
		}
	})
}

type oilData struct {
	Province string  `json:"province"`
	O89      float64 `json:"o89"`
	O92      float64 `json:"o92"`
	O95      float64 `json:"o95"`
	O98      float64 `json:"o98"`
	O0       float64 `json:"o0"`
}
type apiResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data []oilData `json:"data"`
}

func getOilPrice(province string) (*oilData, error) {
	var res apiResponse
	if err := req.C().Get(OilPriceURL).Do().Into(&res); err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, nil
	}
	var data oilData
	for _, p := range res.Data {
		if p.Province == province {
			data = p
		}
	}

	return &data, nil
}
