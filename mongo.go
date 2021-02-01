package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	bson "go.mongodb.org/mongo-driver/bson"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoConfig is a Struct that holds mongo config
type mongoConfig struct {
	authMechanism string
	host          string
	port          int
	user          string
	password      string
	caCert        string
}

type mongoClient struct {
	client *mongo.Client
}

type User struct {
	ID      string   `bson:"userId"`
	Folders []string `bson:"folders"`
}

type MetadataObject struct {
	AccessionID string `bson:"accessionId"`
	Schema      string `bson:"schema"`
}

type MetadataCollection struct {
	FolderID        string           `bson:"folderId"`
	MetadataObjects []MetadataObject `bson:"metadataObjects"`
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

func getAccessionIdsAndSchemas(metadataCollections []MetadataCollection) ([]string, []string) {

	var schemas []string
	var accessionIds []string

	for _, col := range metadataCollections {
		for _, obj := range col.MetadataObjects {
			accessionIds = append(accessionIds, obj.AccessionID)
			schemas = append(schemas, obj.Schema)
		}
	}
	return removeStrDuplicates(accessionIds), removeStrDuplicates(schemas)
}

func removeStrDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	res := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			res = append(res, elements[v])
		}
	}
	return res
}

func (c mongoClient) getUserFolders(database string, collection string, userID string) []string {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"userId": userID}
	users := c.client.Database(database).Collection(collection)
	var user User
	err := users.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		log.Info(err)
	}
	log.Debugf("User %s has folders : %s", user.ID, strings.Join(user.Folders, ","))
	return user.Folders

}

func (c mongoClient) getMetadataObjects(database string, collection string, accessionIds []string) {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)
	filter := bson.M{"accessionId": bson.M{"$in": accessionIds}}
	users := c.client.Database(database).Collection(collection)
	var objects []interface{}
	cursor, err := users.Find(context.TODO(), filter)
	if err != nil {
		log.Info(err)
	}
	err = cursor.All(context.TODO(), &objects)
	if err != nil {
		log.Info(err)
	}
	log.Infof("Objects found in collection %s", collection)
	log.Infoln(strings.Repeat("-", 10))
	log.Info(objects)
	log.Infoln(strings.Repeat("-", 10))

}

func (c mongoClient) getMetadataCollections(database string, collection string, folder []string) []MetadataCollection {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"folderId": bson.M{"$in": folder}}
	users := c.client.Database(database).Collection(collection)
	var mc []MetadataCollection
	cursor, err := users.Find(context.TODO(), filter)
	if err != nil {
		log.Info(err)
	}
	err = cursor.All(context.TODO(), &mc)
	if err != nil {
		log.Info(err)
	}
	//log.Info(mc)
	return mc

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
