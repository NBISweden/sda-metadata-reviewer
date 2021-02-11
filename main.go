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

	inbox, err := newS3Backend(conf.s3)
	if err != nil {
		log.Error(err)
	}
	log.Debug(inbox)

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
	case "cross-ref-inbox":
		{
			log.Info("Cross reference started")
			var analysisAccession string
			if metadataFilter.AccessionID != "" {
				analysisAccession = metadataFilter.AccessionID
			} else if metadataFilter.FolderID != "" {
				analysisAccession = client.getAccessionFromAnalysis("folders", "folder", metadataFilter.FolderID)
			}
			files := client.getFilesFromAnalysis("objects", "analysis", analysisAccession)

			for _, file := range files {
				exists, err := inbox.GetFileSize(file.FileName)
				if err != nil {
					log.Debugf("Error accessing s3: %s", err)
					break
				}
				if exists {
					log.Infof("File %s exists", file.FileName)
				} else {
					log.Infof("File %s does not exist", file.FileName)
				}
			}
		}
	case "cross-ref-ingestion":
		{
			log.Info("Cross reference started")
			postgres, err := NewDB(conf.postgres)
			if err != nil {
				log.Error(err)
			}

			var analysisAccession string
			if metadataFilter.AccessionID != "" {
				analysisAccession = metadataFilter.AccessionID
			} else if metadataFilter.FolderID != "" {
				analysisAccession = client.getAccessionFromAnalysis("folders", "folder", metadataFilter.FolderID)
				log.Infof(analysisAccession)
			}
			files := client.getFilesFromAnalysis("objects", "analysis", analysisAccession)

			for _, file := range files {
				err := postgres.GetChecksum(file)
				if err != nil {
					log.Infof("File %s does not exists", file.FileName)
				} else {
					log.Infof("File %s exists", file.FileName)
				}
			}
		}
	}

	client.disconnectFromMongo()

}
