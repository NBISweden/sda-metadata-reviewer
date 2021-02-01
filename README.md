# Querying collections in the metadata store


## Dependency

```shell
brew install mongodb-community
```
## Start the mongo server

```shell
cd dev_tools/
docker-compose up -d
```
## Exporting a mongo dump

```shell
mongodump -u admin -p admin > metadata.bson
```

## Restoring a mongo dump

```shell
mongorestore -u admin -p admin metadata.bson
```

## Define a filter

```json
{
    "userId": "4153a24c78fb4f4f98cd50f5e05f577c",
    "folderId": "6f936ffa72874241a5d78cc9c8b6ddfe",
    "accessionId": ""
}
```
## Querying the metadata store

```shell
go build
./main 2>/dev/null
```
