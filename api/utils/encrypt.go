package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"sfilter/config"
	"sfilter/utils"
)

func AesEncrypt(data interface{}, key string) (interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	keyBytes := []byte(key) // 将string类型的key转换为[]byte类型
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	msg := pad(jsonData)
	// msg := jsonData
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := make([]byte, aes.BlockSize) // 初始化一个与块大小相同的初始向量

	// 以CBC模式进行加密
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], msg)

	// 将加密后的数据存储在interface{}类型的变量中并返回
	result := interface{}(ciphertext)
	return result, nil
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - (len(src) % aes.BlockSize)
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(src, padtext...)
}

func TEST_ENCRYPT() {
	data := "abcdefg"
	key := config.API_AES_DATA_KEY

	enc, err := AesEncrypt(data, key)
	utils.Tracef("key: %v, enc: %s, err: %v", key, enc, err)
}
