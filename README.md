# Querying collections in the metadata store
## Start the mongo server

```shell
cd dev_tools/
docker-compose up -d
```

## Poupulating a database

```shell
docker exec -it mongo sh

# mongo -u datasteward -p datasteward

> use metadata

switched to db metadata

> db.accessionId.insert({"id": "SEGA6543"})

WriteResult({ "nInserted" : 1 })
```

## Querying the metadata store

```shell
go build
./main --collection "accessionId"
```
