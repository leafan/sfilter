package utils

import (
	"time"

	"github.com/ysicing/workwxbot"
)

var debugMessage = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=3d1a1d63-d4af-449b-a61d-434adb9f14e3"

func SendWecommDebugMsg(content string) {
	SendWecommBot(debugMessage, content)
}

func SendWecommBot(robotUrl, content string) {
	wxbot := workwxbot.NewRobot(robotUrl)

	msg := workwxbot.WxBotMessage{
		MsgType: "markdown",
		MarkDown: workwxbot.BotMarkDown{
			Content: content,
		},
	}

	err := wxbot.Send(msg)
	if err != nil {
		Warnf("[ SendWecommBot ] wxbot.Send error: ", err)
		time.Sleep(time.Second)
	}
}
