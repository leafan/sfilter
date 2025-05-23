.PHONY:  run ps build stop restart bsc creat filter api wiser deepeye dwiser cwiser

default: build

bsc:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_bsc$$' 2>&1 || true
	pkill -f '^.*sapi_bsc$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter_eth /backup/bsc_filter/sfilter_bsc
	cp sapi /backup/bsc_filter/sapi_bsc

	@echo "\n\033[0;34mStart process...\033[0m"

	cd /backup/bsc_filter/ && nohup ./sapi_bsc > /backup/bsc_filter/logs/sapi_$(shell date +%s).log 2>&1 &
	cd /backup/bsc_filter/ && nohup ./sfilter_bsc > /backup/bsc_filter/logs/sfilter_$(shell date +%s).log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"


creat:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_creat$$' 2>&1 || true
	pkill -f '^.*sapi_creat$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter_eth /data_v1/creat/sfilter_creat
	cp sapi /data_v1/creat/sapi_creat

	@echo "\n\033[0;34mStart process...\033[0m"

	cd /data_v1/creat/ && nohup ./sapi_creat > /data_v1/creat/logs/sapi_creat.log 2>&1 &
	cd /data_v1/creat/ && nohup ./sfilter_creat > /data_v1/creat/logs/sfilter_creat.log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"

deepeye:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*sfilter_deepeye$$' 2>&1 || true
	pkill -f '^.*sapi_deepeye$$' 2>&1 || true

	@echo "\n\033[0;34mCopy file...\033[0m"
	cp sfilter_eth /backup/deepeye/sfilter_deepeye
	cp sapi /backup/deepeye/sapi_deepeye

	@echo "\n\033[0;34mStart process...\033[0m"

	# cd /backup/deepeye/ && nohup ./sapi_deepeye > /backup/deepeye/logs/sapi_deepeye_$(shell date +%s).log 2>&1 &
	cd /backup/deepeye/ && nohup ./sfilter_deepeye > /backup/deepeye/logs/sfilter_deepeye_$(shell date +%s).log 2>&1 &

	@echo "\n\033[0;34mFinished...\033[0m"

dwiser:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*swiser_deepeye$$' 2>&1 || true
	sleep 1

	cp swiser /backup/deepeye/swiser_deepeye

	cd /backup/deepeye/ && nohup ./swiser_deepeye -wiser -deal > /backup/deepeye/logs/swiser_deepeye.log 2>&1 &

cwiser:
	@make build

	@echo "\n\033[0;34mKill process...\033[0m"
	pkill -f '^.*swiser_creat$$' 2>&1 || true
	sleep 1

	cp swiser /data_v1/creat/swiser_creat

	cd /data_v1/creat/ && nohup ./swiser_creat -wiser > /data_v1/creat/logs/swiser_creat.log 2>&1 &


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