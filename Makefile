.PHONY: sapi sfilter run ps build stop restart  creat

default: build

creat:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_creat$$' 2>&1 || true
	pkill -f '^.*sapi_creat$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter /data_v1/deepeye/sfilter_creat
	cp sapi /data_v1/deepeye/sapi_creat

	@echo "\n\033[0;34mStart process...\033[0m"

	cd /data_v1/deepeye/ && nohup ./sapi_creat > /var/local/log/sapi_creat.log 2>&1 &
	cd /data_v1/deepeye/ && nohup ./sfilter_creat > /var/local/log/sfilter_creat.log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"

run:
	@make start
	
sapi:
	go run cmd/api/main.go

sfilter:
	go run cmd/sfilter/main.go

restart:
	@make --no-print-directory stop
	@make --no-print-directory start

start:
	@make --no-print-directory build
	
	@echo "\n\033[0;34mChecking sfilter now...\033[0m"
	@if pgrep -f '^.*sfilter$$'; then \
        echo "\033[0;32msfilter process is already running\033[0m"; \
    else \
        echo "\033[0;31mStarting sfilter process...\033[0m"; \
        nohup ./sfilter > /data_v1/logs/sfilter.log 2>&1 & \
    fi

	@echo "\n\033[0;34mChecking sapi now...\033[0m"
	@if pgrep -f '^.*sapi$$'; then \
        echo "\033[0;32msapi process is already running\033[0m"; \
    else \
        echo "\033[0;31mStarting sapi process...\033[0m"; \
        nohup ./sapi > /data_v1/logs/sapi.log 2>&1 & \
    fi

	@echo ""

	@make --no-print-directory ps

stop:
	@echo "\033[0;34mpkill now...\033[0m"
	pkill -f '^.*sfilter$$' 2>&1 || true
	pkill -f '^.*sapi$$' 2>&1 || true

ps:
	@echo "\033[0;34mmake ps result...\033[0m"
	@ps -e | grep -w "sapi\|sfilter" | grep -v "grep" || true

log:
	@tail -f /data_v1/logs/sapi.log 


# 停止所有服务
build:
	@echo "\033[0;34mbuild now...\033[0m"
	go build -o sapi cmd/api/main.go
	go build -o sfilter cmd/sfilter/main.go


