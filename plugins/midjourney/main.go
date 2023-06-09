package midjourney

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/pkg/log"
	"github.com/yqchilde/wxbot/engine/pkg/sqlite"
	"github.com/yqchilde/wxbot/engine/robot"
)

var (
	db sqlite.DB // 数据库
)

func init() {
	engine := control.Register("midjourney", &control.Options{
		Alias: "🌄MidJourney文生图",
		Help: `指令：
1.使用：/mj 提示词
2.查询本人剩余的使用次数：get mymjlimit
购买：（￥0.5/次）
1.添加助手为微信好友
2.向助手转账，备注信息必填：mj
3.助手会自动确认转账信息，然后开通/mj指令权限并返回剩余使用次数`,
	})
	if err := sqlite.Open("data/plugins/midjourney/midjourney.db", &db); err != nil {
		log.Fatalf("open sqlite db failed: %v", err)
	}
	if err := db.Create("token", &Token{}); err != nil {
		log.Fatalf("create token table failed: %v", err)
	}
	if err := db.Create("whitelist", &Whitelist{}); err != nil {
		log.Fatalf("create whitelist table failed: %v", err)
	}
	if err := db.Create("accesslimit", &AccessLimit{}); err != nil {
		log.Fatalf("create accesslimit table failed: %v", err)
	}
	// 设置token
	// 操作权限：管理员
	engine.OnRegex("set mj token (.*)", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		token := ctx.State["regex_matched"].([]string)[1]
		data := Token{Token: token}
		if err := db.Orm.Table("token").Where(&data).FirstOrCreate(&data).Error; err != nil {
			ctx.ReplyText(fmt.Sprintf("token设置失败: %v", token))
			return
		}
		ctx.ReplyText(fmt.Sprintf("token设置成功: %v", token))
	})
	// 删除token
	// 操作权限：管理员
	engine.OnRegex("del mj token", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		if err := db.Orm.Table("token").Where("1 = 1").Delete(&Token{}).Error; err != nil {
			ctx.ReplyText(fmt.Sprintf("token删除失败: %v", err.Error()))
			return
		}
		ctx.ReplyText("token删除成功")
	})
	// 添加白名单
	// 操作权限：管理员
	engine.OnRegex("add mj whitelist (.*)", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		whitelist := strings.Split(ctx.State["regex_matched"].([]string)[1], ";")
		failedOnes := make([]string, 0)
		for i := range whitelist {
			data := Whitelist{WxId: whitelist[i]}
			if err := db.Orm.Table("whitelist").Where(&data).FirstOrCreate(&data).Error; err != nil {
				failedOnes = append(failedOnes, whitelist[i])
				continue
			}
		}
		if len(failedOnes) > 0 {
			ctx.ReplyText(fmt.Sprintf("以下白名单设置失败: %v", failedOnes))
			return
		}
		ctx.ReplyText(fmt.Sprintf("白名单设置成功: %v", strings.Join(whitelist, ",")))
	})
	// 删除白名单
	// 操作权限：管理员
	engine.OnRegex("del mj whitelist (.*)", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		whitelist := strings.Split(ctx.State["regex_matched"].([]string)[1], ";")
		failedOnes := make([]string, 0)
		for i := range whitelist {
			if err := db.Orm.Table("whitelist").Where("wx_id = ?", whitelist[i]).Delete(&Whitelist{}).Error; err != nil {
				failedOnes = append(failedOnes, whitelist[i])
				continue
			}
		}
		if len(failedOnes) > 0 {
			ctx.ReplyText(fmt.Sprintf("以下白名单删除失败: %v", failedOnes))
			return
		}
		ctx.ReplyText("白名单删除成功")
	})
	// 文生图
	// 操作权限：所有人
	engine.OnRegex(`^/mj\s+(.*?)$`).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		prompt := ctx.State["regex_matched"].([]string)[1]

		var accesslimits []AccessLimit
		if err := db.Orm.Table("accesslimit").Where("wx_id = ?", ctx.Event.FromWxId).Find(&accesslimits).Error; err != nil {
			log.Errorf("查询剩余使用次数出错: %v", err)
			return
		}
		if len(accesslimits) == 0 || accesslimits[0].Limit == 0 {
			log.Errorf("使用次数超限，您可通过转账购买mj使用次数\n￥0.5/次")
			ctx.ReplyTextAndAt("使用次数超限，您可通过转账购买mj使用次数\n￥0.5/次")
			return
		}
		var accesslimit AccessLimit
		accesslimit.Limit = accesslimits[0].Limit - 1
		accesslimit.WxId = accesslimits[0].WxId

		// var whitelist []Whitelist
		// if err := db.Orm.Table("whitelist").Where("wx_id=?", ctx.Event.FromWxId).Find(&whitelist).Error; err != nil {
		// 	ctx.ReplyTextAndAt("👎👎👎\n获取白名单出错")
		// 	return
		// }

		// if len(whitelist) == 0 {
		// 	ctx.ReplyTextAndAt("🔴🔴🔴\n暂无文生图权限，联系管理员开通！")
		// 	return
		// }

		var reqData ApiRequest
		var createFlag bool
		var actionStr string
		re := regexp.MustCompile(`(\d+) (u\d+|v\d+)`)
		if match := re.FindStringSubmatch(prompt); match == nil {
			//生成
			createFlag = true
			ctx.ReplyTextAndAt("🚀🚀🚀\n图片生成中，请等候...")
			reqData = ApiRequest{
				Prompt: prompt,
			}
		} else {
			//调整
			createFlag = false
			ctx.ReplyTextAndAt(fmt.Sprintf("🚀🚀🚀\n正在调整图片%s，请等候...", match[1]))
			if actionStr, err := getAction(match[2]); err == nil {
				reqData = ApiRequest{
					Prompt:  prompt,
					ImageId: match[1],
					Action:  actionStr,
				}
			} else {
				ctx.ReplyTextAndAt("👎👎👎\n参数输入有误，请检查后重新输入")
				return
			}
		}
		if data, err := imagine(reqData); err == nil {
			if data == nil {
				ctx.ReplyTextAndAt("👎👎👎\n图片生成失败，一定是哪里出错了")
			} else {
				if err := db.Orm.Table("accesslimit").Where("wx_id", ctx.Event.FromWxId).Save(&accesslimit).Error; err != nil {
					fmt.Println("更新剩余使用次数失败", err)
				}

				if createFlag {
					ctx.ReplyTextAndAt(fmt.Sprintf(`👍👍👍
%s
👍👍👍
🎽剩余使用次数：%f
🥇ImageId:%s
🏸放大：请用 v+ImageId来告诉我您喜欢哪一张。例如，u1，我将会根据您的选择画出第一张的更精美的版本。
💥变换：如果您对所有的草图都不太满意，但是对其中某一张构图还可以，可以用 v+ImageId来告诉我，我会画出类似的四幅图供您选择。
✍️具体操作：/mj ImageId 操作，比如 /mj %s u1`, prompt, accesslimit.Limit, data.ImageId, data.ImageId))
				} else {
					ctx.ReplyTextAndAt(fmt.Sprintf(`👍👍👍
%s %s
👍👍👍
🎽剩余使用次数：%.2f`, data.ImageId, actionStr, accesslimit.Limit))
				}
				ctx.ReplyImage("local://" + data.ImageLocalPath)
			}
		} else {
			ctx.ReplyTextAndAt(fmt.Sprintf("👎👎👎\n图片生成失败[%s]", err.Error()))
		}

	})

	//收款确认
	engine.OnMessage().SetBlock(false).Handle(func(ctx *robot.Ctx) {
		// 监听转账事件
		if ctx.IsEventTransfer() {
			f := ctx.Event.TransferMessage

			// 排除其他类型转账信息，例如退回转账也会触发事件
			if f.Memo == "mj" && f.MsgSource == 1 {
				if err := ctx.ConfirmTransfer(f.FromWxId, f.TransferId); err == nil {

					// 将字符串金额转换为浮点数
					money, err := strconv.ParseFloat(f.Money, 64)
					if err != nil {
						fmt.Println("无法解析数字：", err)
						return
					}

					// 计算 money/0.5 并取整
					limit := math.Round(money / 0.5)

					var accesslimit AccessLimit
					accesslimit.WxId = f.FromWxId

					var accesslimits []AccessLimit
					if err := db.Orm.Table("accesslimit").Where("wx_id = ?", f.FromWxId).Find(&accesslimits).Error; err != nil {
						log.Errorf("查询剩余使用次数出错: %v", err)
						return
					}

					if len(accesslimits) == 0 {
						accesslimit.Limit = limit
						if err := db.Orm.Table("accesslimit").Create(&accesslimit).Error; err != nil {
							fmt.Println("保存次数出错", err)
							return
						}
					} else {
						accesslimit.Limit = limit + accesslimits[0].Limit
						if err := db.Orm.Table("accesslimit").Where("wx_id", f.FromWxId).Save(&accesslimit).Error; err != nil {
							fmt.Println("新增次数出错", err)
							return
						}
					}

					resp := fmt.Sprintf("确认收款：￥%s\n剩余mj使用次数：%.2f", f.Money, accesslimit.Limit)

					log.Printf(resp)
					ctx.SendText(f.FromWxId, resp)
					return
				} else {
					log.Errorf("确认收款请求失败: %v", err)
					return
				}
			}

		}
	})

	// 查询剩余次数
	// 操作权限：管理员
	engine.OnRegex("get mjlimit (.*)", robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		wx_id := args[1]
		log.Printf(wx_id)
		var accesslimits []AccessLimit
		if err := db.Orm.Table("accesslimit").Where("wx_id = ?", wx_id).Find(&accesslimits).Error; err != nil {
			log.Errorf("查询剩余使用次数出错: %v", err)
			return
		}
		if len(accesslimits) == 0 {
			ctx.ReplyText("该用户尚未购买")
		} else {
			ctx.ReplyText(fmt.Sprintf(`该用户的剩余使用次数为：%.2f`, accesslimits[0].Limit))
		}
	})
	// 查询本人剩余次数
	// 操作权限：所有人
	engine.OnRegex("get mymjlimit").SetBlock(true).Handle(func(ctx *robot.Ctx) {
		var accesslimits []AccessLimit
		if err := db.Orm.Table("accesslimit").Where("wx_id = ?", ctx.Event.FromWxId).Find(&accesslimits).Error; err != nil {
			log.Errorf("查询剩余使用次数出错: %v", err)
			return
		}
		if len(accesslimits) == 0 {
			ctx.ReplyTextAndAt("您尚未购买")
		} else {
			ctx.ReplyTextAndAt(fmt.Sprintf(`您的剩余使用次数为：%.2f`, accesslimits[0].Limit))
		}
	})
	// 手动添加使用次数
	// 操作权限：管理员
	engine.OnRegex(`add mjlimit (.*) (\d+)`, robot.OnlyPrivate, robot.AdminPermission).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		wx_id, limit_to_add_str := args[1], args[2]
		limit_to_add, err := strconv.ParseFloat(limit_to_add_str, 64)
		if err != nil {
			log.Errorf("无法将字符串转换为浮点数：%v", err)
			return
		}
		var accesslimits []AccessLimit
		var accesslimit AccessLimit
		if err := db.Orm.Table("accesslimit").Where("wx_id = ?", wx_id).Find(&accesslimits).Error; err != nil {
			log.Errorf("查询剩余使用次数出错: %v", err)
			return
		}
		if len(accesslimits) == 0 {
			accesslimit.WxId = wx_id
			accesslimit.Limit = limit_to_add
			if err := db.Orm.Table("accesslimit").Create(&accesslimit).Error; err != nil {
				log.Errorf("保存次数出错：%v", err)
				return
			}
		} else {
			accesslimit = accesslimits[0]
			accesslimit.Limit = accesslimit.Limit + limit_to_add
			if err := db.Orm.Table("accesslimit").Where("wx_id", wx_id).Save(&accesslimit).Error; err != nil {
				log.Errorf("新增次数出错：%v", err)
				return
			}
		}
		ctx.ReplyText(fmt.Sprintf(`新增成功！
当前剩余次数：%.2f`,
			accesslimit.Limit))
	})
}

func getAction(val string) (string, error) {
	re1 := regexp.MustCompile(`u(\d+)`)
	if match1 := re1.FindStringSubmatch(val); match1 != nil {
		return "upsample" + match1[1], nil
	}
	re2 := regexp.MustCompile(`v(\d+)`)
	if match2 := re2.FindStringSubmatch(val); match2 != nil {
		return "variation" + match2[1], nil
	}
	return "", nil
}
