package ocr

// 调用本地的python脚本
// pip install backgroundremover
// backgroundremover -i 11.png -o 22.png
// 第一次运行命令需要科学上网下载模型包

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/yqchilde/wxbot/engine/control"
	"github.com/yqchilde/wxbot/engine/robot"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// simplifiedchinese.GBK.NewEncoder().Bytes()   //utf-8 转 gbk
// simplifiedchinese.GBK.NewDecoder().Bytes()  //gbk 转 utf-8

func init() {
	engine := control.Register("ocr", &control.Options{
		Alias: "🔠OCR文字识别",
		Help: `指令：ocr|文字识别
功能：一键实现OCR文字识别
使用：发送指令“ocr”或“文字识别”，然后在30秒内发送要处理的图片`,
	})
	engine.OnRegex(`(^ocr|^OCR|^文字识别)$`, robot.MustPicture).SetBlock(true).Handle(func(ctx *robot.Ctx) {
		imageUrl := ctx.State["image_url"].(string)
		ctx.ReplyTextAndAt("🕓🕓🕓努力识别....")
		cwd, err := os.Getwd()
		if err != nil {
			return
		}
		script := path.Join(cwd, "plugins/ocr/ocr.py")
		cmd := exec.Command("python", script, imageUrl)

		log.Println(cmd)
		output, err := cmd.CombinedOutput()
		output, err = simplifiedchinese.GBK.NewDecoder().Bytes(output) //如果有中文需要转化为utf8
		if err != nil {
			log.Panicln(err)
			ctx.ReplyTextAndAt(fmt.Sprintf("🅰️识别失败[%s %v]", output, err))
			return
		}
		re := regexp.MustCompile(`result===>(.*)`)
		match := re.FindStringSubmatch(string(output))
		if len(match) > 0 {
			fmt.Println(match[1])
			ctx.ReplyTextAndAt(match[1])
		}
	})
}
