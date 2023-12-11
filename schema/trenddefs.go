package schema

import "time"

// 趋势线结构体元结构
type TrendElem struct {
	Value float64   `json:"value" bson:"value"`
	Time  time.Time `json:"time" bson:"time"`
}

type TrendStruct []TrendElem
