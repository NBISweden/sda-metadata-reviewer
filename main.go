package main

import (
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2/json"
)

// List all folder names together with their ids and user ids
// JSON output on stdout

type metadataFilter struct {
	UserID      string `json:"userId"`
	FolderID    string `json:"folderId"`
	AccessionID string `json:"accessionId"`
}

func main() {

	conf := NewConfig()

	client, err := newMongoClient(conf.mongo)

	if err != nil {
		log.Error(err)
	}

	jsonFilter, err := ioutil.ReadFile("filter.json")
	if err != nil {
		log.Error(err)
	}
	//jsonFilter := `{"userId": "4153a24c78fb4f4f98cd50f5e05f577c", "folderId": "76566fb39a7644a7957940ffca0aca99", "accessionId": "06c56cbc9290475397f169db800155c5"}`
	log.Debug(string(jsonFilter))

	metadataFilter := metadataFilter{}
	err = json.Unmarshal([]byte(jsonFilter), &metadataFilter)

	if err != nil {
		log.Error(err)
	}

	client.connectToMongo()

	var userFolders []string

	if metadataFilter.FolderID != "" {
		userFolders = append(userFolders, metadataFilter.FolderID)
	} else {
		userFolders = client.getUserFolders("users", "user", metadataFilter.UserID)
	}

	metadataCollections := client.getMetadataCollections("folders", "folder", userFolders)

	var accessionIds []string
	var schemas []string

	if metadataFilter.AccessionID != "" {
		accessionIds = append(accessionIds, metadataFilter.AccessionID)
		_, schemas = getAccessionIdsAndSchemas(metadataCollections)
	} else {
		accessionIds, schemas = getAccessionIdsAndSchemas(metadataCollections)
	}

	log.Debugf("Accession ids are: %s", strings.Join(accessionIds, " "))
	log.Debugf("Schemas are: %s", strings.Join(schemas, " "))

	for _, sch := range schemas {
		client.getMetadataObjects("objects", sch, accessionIds)
	}
	client.disconnectFromMongo()

}
