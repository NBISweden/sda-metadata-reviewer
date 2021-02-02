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
	log.Debug(string(jsonFilter))

	metadataFilter := metadataFilter{}
	err = json.Unmarshal(jsonFilter, &metadataFilter)

	if err != nil {
		log.Error(err)
	}

	client.connectToMongo()

	action := getCLflags().action

	switch action {
	case "list-folders":
		{
			user := client.getUser("users", "user", metadataFilter.UserID)
			client.getFolders("folders", "folder", user.Folders)
		}
	case "list-objects":
		{
			var userFolders []string

			if metadataFilter.FolderID != "" {
				userFolders = append(userFolders, metadataFilter.FolderID)
			} else {
				userFolders = client.getUser("users", "user", metadataFilter.UserID).Folders
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
		}
	case "list-users":
		{
			client.getAllUsers("users", "user")
		}

	}
	client.disconnectFromMongo()

}
