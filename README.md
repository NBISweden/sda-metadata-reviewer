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
mongodump -u admin -p admin
```

## Restoring a mongo dump

```shell
mongorestore -u admin -p admin metadata.bson
```

## Reviewing the submission metadata

 Minimally the username needs to be provided in the `filter.json` file to fetch the folder identifiers that belong to a user's submission folder. To obtain them, fill in the username in the aforementioned file:

```json
{
    "username": "test test"
}
```

```shell
./main --action list-folders
```

Now find the folder id of the submission and specify it in the `filter.json` file:

```json
{
    "folderId": "d28e77a17a6a4c19ac53891a678054a5"
}
```

- In order to fetch all user specific metadata objects from a given folder run:

```shell
./main --action list-objects
```

* If you only want to see the metadata from a given metadata object, it is possible to specify its accessionId as a filter.

```json
{
    "folderId": "d28e77a17a6a4c19ac53891a678054a5",
    "accessionId": ""
}
```
