package transferconfirm

import (
	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/robot"
)

func init() {
	engine := control.Register("transferconfirm", &control.Options{
		Alias:    "自动确认收款",
		HideMenu: true,
	})

	engine.OnMessage().SetBlock(false).Handle(func(ctx *robot.Ctx) {
		// 监听转账事件
		// if ctx.IsEventTransfer() {
		// 	f := ctx.Event.TransferMessage

		// 	if f.Memo == "mj" {
		// 		if err := ctx.ConfirmTransfer(f.FromWxId, f.TransferId); err == nil {

		// 			resp := fmt.Sprintf("确认收款：￥%s，获得mj使用次数：%s", f.Money, f.Memo)

		// 			log.Printf(resp)

		// 			// 根据收款金额，计算用户的mj使用次数
		// 			ctx.SendText(f.FromWxId, resp)
		// 			return
		// 		} else {
		// 			log.Errorf("确认收款请求失败: %v", err)
		// 			return
		// 		}
		// 	}

		// }
	})
}
