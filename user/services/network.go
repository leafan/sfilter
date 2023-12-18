package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sfilter/user/config"
)

func GetIpLocation(ip string) (string, error) {
	var location = "Unknown"

	url := fmt.Sprintf("https://ipinfo.io/%s/json?token=%s", ip, config.IPINFO_APIKEY)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return location, err
	}

	// 模拟浏览器请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return location, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return location, err
	}

	city := data["city"].(string)
	region := data["region"].(string)
	country := data["country"].(string)

	location = fmt.Sprintf("%v, %v, %v", city, region, country)

	// utils.Tracef("[ GetIpLocation ] ip: %v, location: %v", ip, location)

	return location, nil
}
