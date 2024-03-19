package utils

import (
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
		Debugf("[ SendWecommBot ] wxbot.Send error: %v", err)
		// time.Sleep(time.Second * 2)
	}

	// time.Sleep(time.Millisecond * 500)
}
