.PHONY: api run

default: build

#查看服务状态
api:
	go run cmd/api/main.go

run:
	go run cmd/sfilter/main.go


# 停止所有服务
build:
	go build -o sapi cmd/api/main.go
	go build -o sfilter cmd/sfilter/main.go


