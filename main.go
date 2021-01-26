package main

import (
	"github.com/prometheus/common/log"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {

	flags := getCLflags()

	//Defaults to study
	collection := flags.collection
	db := flags.db

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
	client.getMetadataObject(db, collection, filter)
	client.disconnectFromMongo()

}
