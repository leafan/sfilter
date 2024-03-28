#!/bin/bash

# created via chatgpt3.5 with modification by leafan

# mongodump / mongorestore 备份和恢复mongo 数据
# 可以定时执行备份

# tips: mongodump 仅备份数据库中的文档, 不备份索引. 所以我们还原后, 需要重新生成索引

# Set the MongoDB host and port
HOST="192.168.2.103"
PORT=27017
DB="eth"
COLLECTIONS="token,pair,hrank,kline1min,kline1hour"

# Backup directory
BACKUP_DIR=/backup/mongodb

# 共保存两份, 有新的覆盖旧的
if [ -d "$BACKUP_DIR/data1" ]; then
    # Check if data2.gz exists
    if [ -d "$BACKUP_DIR/data2" ]; then
        # Delete the older file 
        rm -rf $BACKUP_DIR/data2
    fi

    # Rename data1.gz to data2
    mv $BACKUP_DIR/data1 $BACKUP_DIR/data2
fi

# 本次保存路径
BACKUP_PATH=$BACKUP_DIR/data1

mkdir -p $BACKUP_PATH
for i in $(echo $COLLECTIONS | tr "," "\n")
    do
        mongodump --host $HOST --port $PORT --db $DB --collection $i  --out $BACKUP_PATH
    done

