package utils

import (
	"fmt"
	"log"
)

const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"

	ColorReset = "\n\033[0m"
)

// 跟踪日志, 打印完就会删除
func Tracef(format string, args ...interface{}) {
	fmt.Printf(ColorYellow+format+ColorReset, args...)
}

// 打印普通日志
func Debugf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func Infof(format string, args ...interface{}) {
	log.Printf(ColorGreen+format+ColorReset, args...)
}

func Warnf(format string, args ...interface{}) {
	format = "[ warn ] " + format
	log.Printf(ColorYellow+format+ColorReset, args...)
}

func Errorf(format string, args ...interface{}) {
	format = "[ error ] " + format
	log.Printf(ColorRed+format+ColorReset, args...)

	// 告警
	content := fmt.Sprintf(format, args...)
	SendWecommDebugMsg(content)
}

// 这个函数会退出
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(ColorRed+format+ColorReset, args...)

	// 告警
	content := fmt.Sprintf(format, args...)
	SendWecommDebugMsg(content)
}
