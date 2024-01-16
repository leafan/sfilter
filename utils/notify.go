package utils

import (
	"os"

	"github.com/ysicing/workwxbot"
)

var debugMessage = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=3d1a1d63-d4af-449b-a61d-434adb9f14e3"

func SendWecommDebugMsg(content string) {
	SendWecommBot(debugMessage, content)
}

func SendWecommBot(robotUrl, content string) {
	wxbot := workwxbot.NewRobot(robotUrl)

	programName := os.Args[0] // 获取程序名称
	msg := workwxbot.WxBotMessage{
		MsgType: "text",
		BotText: workwxbot.BotText{
			Content: programName + ": " + content,
		},
	}

	err := wxbot.Send(msg)
	if err != nil {
		Warnf("[ SendWecommBot ] wxbot.Send error: ", err)
	}
}
