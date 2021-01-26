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
    "userId": "user001"
}
```
## Querying the metadata store

```shell
go build
./main --db "objects" --collection "study"
```
