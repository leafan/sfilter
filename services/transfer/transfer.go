package transfer

import (
	"context"
	"log"
	"sfilter/config"
	"sfilter/schema"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func SaveTransferEvent(_transfer *schema.Transfer, mongodb *mongo.Client) {
	collection := mongodb.Database(config.DatabaseName).Collection(config.TransferTableName)

	_transfer.CreatedAt = time.Now()
	_, err := collection.InsertOne(context.Background(), _transfer)
	if err != nil {
		log.Printf("[ SaveTransferEvent ] InsertOne error: %v, trasnfer: %v, LogIndexWithTx: %v\n", err, _transfer, _transfer.LogIndexWithTx)
	}
}
