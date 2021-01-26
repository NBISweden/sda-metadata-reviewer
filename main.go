package main

import (
	"io/ioutil"

	"github.com/prometheus/common/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {

	flags := getCLflags()

	//Defaults to study
	collection := flags.collection
	//Defaults to objects
	db := flags.db

	conf := NewConfig()

	client, err := newMongoClient(conf.mongo)

	if err != nil {
		log.Info(err)
	}

	jsonFilter, err := ioutil.ReadFile("filter.json")
	if err != nil {
		log.Info(err)
	}
	var doc interface{}
	err = bson.UnmarshalExtJSON(jsonFilter, false, &doc)
	if err != nil {
		log.Info(err)
	}

	client.connectToMongo()
	client.getMetadataObject(db, collection, doc.(primitive.D))
	client.disconnectFromMongo()

}
