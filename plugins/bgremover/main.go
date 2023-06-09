package bgremover

// 调用本地的backgroundremover命令实现抠图功能
// pip install backgroundremover
// backgroundremover -i 11.png -o 22.png
// 第一次运行命令需要科学上网下载模型包

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/robot"
)

func init() {
	engine := control.Register("bgremover", &control.Options{
		Alias: "⭕AI抠图",
		Help: `指令：抠图
功能：一键抠图，头像背景、身份证背景等使用AI模型一键处理成透明
使用：发送指令“抠图”，然后在30秒内发送要处理的图片`,
	})

	engine.OnFullMatch("抠图", robot.MustPicture).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		imageUrl := ctx.State["image_url"].(string)
		fileName := filepath.Base(imageUrl)
		ctx.ReplyTextAndAt("⛏️⛏️⛏️抠啊抠啊抠....")
		cwd, err := os.Getwd()
		if err != nil {
			return
		}
		out := path.Join(cwd, "data/plugins/bgremover/out", fileName)
		cmd := exec.Command("backgroundremover", "-i", imageUrl, "-o", out)

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Panicln(err)
			ctx.ReplyTextAndAt(fmt.Sprintf("🅰️抠图失败[%s %v]", output, err))
			return
		}
		ctx.ReplyImage(out)
	})
}
