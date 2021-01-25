package main

import (
	"github.com/prometheus/common/log"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {

	flags := getCLflags()

	//Defaults to accessionId
	collection := flags.collection

	conf := NewConfig()

	client, err := newMongoClient(conf.mongo)

	if err != nil {
		log.Info(err)
	}

	// -- BSON object examples --
	// filter := bson.D{{"id", "SEGA6543"}}
	// filter := bson.D{{"id", "SEGA6543"}, {"owner", "user1234"}}

	filter := bson.D{}
	client.connectToMongo()
	client.getMetadataObject(conf.mongo.database, collection, filter)
	client.disconnectFromMongo()

}
