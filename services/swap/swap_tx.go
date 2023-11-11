package swap

import (
	"context"
	"log"
	"sfilter/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

func SaveSwapTx(swap *schema.Swap, mongodb *mongo.Client) {
	collection := mongodb.Database("deepeye").Collection("swap")

	_, err := collection.InsertOne(context.Background(), swap)
	if err != nil {
		log.Printf("[ saveSwapTx ] InsertOne error: %v, swap tx: %v\n", err, swap.Tx)
		return
	}

}
