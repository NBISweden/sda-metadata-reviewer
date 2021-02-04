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
	Name    string   `bson:"name"`
	Eppn    string   `bson:"eppn"`
	Folders []string `bson:"folders"`
}

type Folder struct {
	ID   string `bson:"folderId"`
	Name string `bson:"name"`
}

type MetadataObject struct {
	AccessionID string `bson:"accessionId"`
	Schema      string `bson:"schema"`
	Files       []File `bson:"files"`
}

type MetadataCollection struct {
	FolderID        string           `bson:"folderId"`
	MetadataObjects []MetadataObject `bson:"metadataObjects"`
}

type File struct {
	FileName       string `bson:"filename"`
	ChecksumMethod string `bson:"checksumMethod"`
	Checksum       string `bson:"checksum"`
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

func (c mongoClient) getFolders(database string, collection string, folderIds []string) {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	users := c.client.Database(database).Collection(collection)
	filter := bson.M{"folderId": bson.M{"$in": folderIds}}
	var folders []Folder
	cursor, err := users.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &folders)
	if err != nil {
		log.Error(err)
	}
	for _, obj := range folders {
		out, err := bson.MarshalExtJSON(obj, false, false)
		if err != nil {
			log.Error(err)
		}
		fmt.Println(string(out))
		fmt.Println(strings.Repeat("-", 10))
	}
}

func (c mongoClient) getUser(database string, collection string, userID string) User {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"userId": userID}
	users := c.client.Database(database).Collection(collection)
	var user User
	err := users.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		log.Error(err)
	}
	out, err := bson.MarshalExtJSON(user, false, false)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(string(out))
	return user

}

func (c mongoClient) getAllUsers(database string, collection string) {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{}
	col := c.client.Database(database).Collection(collection)
	var users []User
	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &users)
	if err != nil {
		log.Error(err)
	}
	for _, usr := range users {
		out, err := bson.MarshalExtJSON(usr, false, false)
		if err != nil {
			log.Error(err)
		}
		fmt.Println(string(out))
		fmt.Println(strings.Repeat("-", 10))
	}

}

func (c mongoClient) getMetadataObjects(database string, collection string, accessionIds []string) {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"accessionId": bson.M{"$in": accessionIds}}
	users := c.client.Database(database).Collection(collection)
	var objects []interface{}
	cursor, err := users.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &objects)
	if err != nil {
		log.Error(err)
	}
	for _, obj := range objects {
		out, err := bson.MarshalExtJSON(obj, false, false)
		if err != nil {
			log.Error(err)
		}
		log.Infof("Objects found in collection %s", collection)
		fmt.Println(string(out))
		fmt.Println(strings.Repeat("-", 10))
	}

}

func (c mongoClient) getMetadataCollections(database string, collection string, folder []string) []MetadataCollection {

	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"folderId": bson.M{"$in": folder}}
	users := c.client.Database(database).Collection(collection)
	var mc []MetadataCollection
	cursor, err := users.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &mc)
	if err != nil {
		log.Error(err)
	}
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

func (c mongoClient) getFilesFromAnalysis(database string, collection string, accessionID string) (files []string) {
	//var files []string
	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"accessionId": accessionID}
	client := c.client.Database(database).Collection(collection)
	objects := []MetadataObject{}
	cursor, err := client.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &objects)
	if err != nil {
		log.Error(err)
	}
	//log.Infof("Objects %s", objects)

	for _, obj := range objects {
		for _, file := range obj.Files {
			files = append(files, file.FileName)
		}
	}

	return files

}

func (c mongoClient) getAccessionFromAnalysis(database string, collection string, folderID string) string {
	var accession string
	log.Debugf("Database %s is being queried using the %s collection", database, collection)

	filter := bson.M{"folderId": folderID}
	client := c.client.Database(database).Collection(collection)
	objects := []MetadataCollection{}
	//ar objects []interface{}
	cursor, err := client.Find(context.TODO(), filter)
	if err != nil {
		log.Error(err)
	}
	err = cursor.All(context.TODO(), &objects)
	if err != nil {
		log.Error(err)
	}

	for _, obj := range objects {
		for _, schema := range obj.MetadataObjects {
			if schema.Schema == "analysis" {
				accession = schema.AccessionID
			}
		}
	}

	return accession

}
