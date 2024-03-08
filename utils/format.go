package utils

import (
	"fmt"
	"time"
)

func ReadibleDuration(duration time.Duration) string {
	durationStr := ""

	hours := int(duration.Hours())
	days := hours / 24
	hours = hours % 24
	minutes := int(duration.Minutes()) % 60

	if days > 30 {
		durationStr = "> 1 month"
	} else {
		durationStr = fmt.Sprintf("%dd %dh %dm\n", days, hours, minutes)
	}

	return durationStr
}

func formatNumber(num float64, unit string) string {
	return fmt.Sprintf("%.2f%s", num, unit)
}

func HumanizeNumber(num float64) string {
	if num >= 1e12 {
		return formatNumber(num/1e12, "T")
	}

	if num >= 1e9 {
		return formatNumber(num/1e9, "G")
	}

	if num >= 1e6 {
		return formatNumber(num/1e6, "M")
	}

	if num >= 1e3 {
		return formatNumber(num/1e3, "K")
	}

	return formatNumber(num, "")
}
