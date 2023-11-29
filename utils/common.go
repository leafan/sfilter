package utils

import (
	"bytes"
	"encoding/gob"
	"sfilter/config"
)

func CheckExistString(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

// 注意: src dst需要用指针
func DeepCopy(src, dst interface{}) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(src); err != nil {
		return err
	}

	return gob.NewDecoder(&buffer).Decode(dst)
}

// 计算百分比
func CalcChange(now, last float64) float32 {
	if last == 0 {
		return config.INFINITE_CHANGE
	}

	delta := now - last

	return float32(delta / now)
}
