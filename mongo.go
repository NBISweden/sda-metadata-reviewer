package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoConfig is a Struct that holds mongo config
type mongoConfig struct {
	authMechanism string
	database      string
	host          string
	port          int
	user          string
	password      string
	caCert        string
}

type mongoClient struct {
	client *mongo.Client
}

func newMongoClient(config mongoConfig) (*mongoClient, error) {

	opts := options.Client()
	//tlsConf := transportConfigMongo(config)

	//opts.SetTLSConfig(tlsConf)
	opts.SetConnectTimeout(time.Second * 10)
	opts.SetAuth(options.Credential{AuthMechanism: config.authMechanism, Username: config.user, Password: config.password})
	opts.ApplyURI(fmt.Sprintf("%s:%d", config.host, config.port))

	client, err := mongo.NewClient(opts)

	return &mongoClient{client: client}, err
}

func (c mongoClient) connectToMongo() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := c.client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure the connection was established
	err = c.client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Connection established to metadata store")

}

func (c mongoClient) disconnectFromMongo() {
	err := c.client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Connection closed to metadata store")

}

func (c mongoClient) getMetadataObject(database string, collection string, filter bson.D) {

	log.Infof("Database being queried is %s", database)
	log.Infof("Looking up the collection name %s", collection)

	metadata := c.client.Database(database).Collection(collection)
	cursor, err := metadata.Find(context.TODO(), filter)

	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var val interface{}

		if err = cursor.Decode(&val); err != nil {
			log.Fatal(err)
		}
		fmt.Println(val)
	}
}

// transportConfigMongo is a helper method to setup TLS for the Mongo client.
func transportConfigMongo(config mongoConfig) *tls.Config {
	cfg := new(tls.Config)

	// Enforce TLS1.2 or higher
	cfg.MinVersion = 2

	// Read system CAs
	var systemCAs, _ = x509.SystemCertPool()
	if reflect.DeepEqual(systemCAs, x509.NewCertPool()) {
		log.Debug("creating new CApool")
		systemCAs = x509.NewCertPool()
	}
	cfg.RootCAs = systemCAs

	if config.caCert != "" {
		cacert, e := ioutil.ReadFile(config.caCert)
		if e != nil {
			log.Fatalf("failed to append %q to RootCAs: %v", cacert, e)
		}
		if ok := cfg.RootCAs.AppendCertsFromPEM(cacert); !ok {
			log.Debug("no certs appended, using system certs only")
		}
	}

	return cfg
}
