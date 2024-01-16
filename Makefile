.PHONY:  run ps build stop restart creat filter api wiser deepeye

default: build

creat:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_creat$$' 2>&1 || true
	pkill -f '^.*sapi_creat$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter_eth /data_v1/deepeye/sfilter_creat
	cp sapi /data_v1/deepeye/sapi_creat

	@echo "\n\033[0;34mStart process...\033[0m"

	# cd /data_v1/deepeye/ && nohup ./sapi_creat > /data_v1/deepeye/logs/sapi_creat_$(shell date +%s).log 2>&1 &
	cd /data_v1/deepeye/ && nohup ./sfilter_creat > /data_v1/deepeye/logs/sfilter_creat_$(shell date +%s).log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"

deepeye:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_deepeye$$' 2>&1 || true
	pkill -f '^.*sapi_deepeye$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter_eth /root/deepeye/sfilter_deepeye
	cp sapi /root/deepeye/sapi_deepeye
	cp swiser /root/deepeye/swiser_deepeye

	@echo "\n\033[0;34mStart process...\033[0m"

	cd /root/deepeye/ && nohup ./sapi_deepeye > /root/deepeye/logs/sapi_deepeye_$(shell date +%s).log 2>&1 &
	cd /root/deepeye/ && nohup ./sfilter_deepeye > /root/deepeye/logs/sfilter_deepeye_$(shell date +%s).log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"


run:
	@make start
	
api:
	pkill -f '^.*sapi$$' 2>&1 || true
	go run cmd/api/main.go $(ARGS)

filter:
	pkill -f '^.*sfilter_eth$$' 2>&1 || true
	go run cmd/sfilter/main.go $(ARGS)

wiser:
	pkill -f '^.*swiser_eth$$' 2>&1 || true
	go run cmd/wiser/main.go $(ARGS)


restart:
	@make --no-print-directory stop
	@make --no-print-directory start

start:
	@make --no-print-directory build
	
	@echo "\n\033[0;34mChecking sfilter_eth now...\033[0m"
	@if pgrep -f '^.*sfilter_eth$$'; then \
        echo "\033[0;32msfilter_eth process is already running\033[0m"; \
    else \
        echo "\033[0;31mStarting sfilter_eth process...\033[0m"; \
        nohup ./sfilter_eth > /data_v1/logs/sfilter_eth.log 2>&1 & \
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
	pkill -f '^.*sfilter_eth$$' 2>&1 || true
	pkill -f '^.*sapi$$' 2>&1 || true

test:
	go run cmd/test/test.go

ps:
	@echo "\033[0;34mmake ps result...\033[0m"
	@ps -e | grep -w "sapi\|sfilter_eth" | grep -v "grep" || true

log:
	@tail -f /data_v1/logs/sapi.log 


build:
	@echo "\033[0;34mbuild now...\033[0m"

	go build -o sapi cmd/api/main.go
	go build -o swiser cmd/wiser/main.go

	@make eth

eth:
	go build -o sfilter_eth cmd/sfilter/main.go