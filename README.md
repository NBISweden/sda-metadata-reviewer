# Querying collections in the metadata store


## Dependency

```shell
brew install mongodb-community
```
## Start the mongo and s3 servers

```shell
cd dev_tools/
docker-compose up -d
```
* The minio is used only for the cross reference, you can skip it using
```shell
cd dev_tools/
docker-compose up -d database
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

- Find the user that you want to review the submission for:

```shell
./main --action list-users
```

 Minimally the user id needs to be provided in the `filter.json` file to fetch the folder identifiers that belong to a user's submission folder. To obtain them, fill in the user id in the aforementioned file:

```json
{
    "userId": "myuserid"
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

## Cross reference metadata files with S3
The cross reference part is comparing the files in the metadata with the ones uploaded in the S3 backend.
It can be run with the `folderId` for the specific submission using the following filter:
```json
{
    "folderId": "d28e77a17a6a4c19ac53891a678054a5",
    "accessionId": ""
}
```
or with the `accessionId` of the `analysis` using the following filter:
```json
{
    "folderId": "",
    "accessionId": "9fd29e35a82e49d999528a5f3c6d49aa"
}
```
to cross reference based on the file names run the following command:
```shell
./main --action cross-ref-inbox
```
to cross reference based on the checksums of the files run the following command:
```shell
./main --action cross-ref-ingestion
```

- Fix docker-compose for postgres and s3