package main

import "github.com/prometheus/common/log"

func main() {

	//flags := getCLflags()

	conf := NewConfig()

	client, err := newMongoClient(conf.mongo)

	if err != nil {
		log.Info(err)
	}

	client.connectToMongo()
	client.getMetadataObject(conf.mongo.database, "submission")
	client.disconnectFromMongo()

}
