# Querying collections in the metadata store
## Start the mongo server

```shell
cd dev_tools/
docker-compose up -d
```

## Poupulating a database

```shell
docker exec -it mongo sh

# mongo -u admin -p admin

> use objects
> db.study.insert({"id": "1", "drafts" : [], "folders": ["4dd7cbac927e4da48fc2a37cb9965a22", "6e84b03667ba4d5ca5f02d4c87834ce0", "7039a8ed04184f4798f47687c027b13a"], "userId": "user001", "name": "se123", "eppn": "test"})
> db.study.insert({"id": "1", "drafts" : [], "folders": ["4dd7cbac927e4da48fc2a37cb9965a22", "6e84b03667ba4d5ca5f02d4c87834ce0", "7039a8ed04184f4798f47687c027b13a"], "userId": "user002", "name": "se124", "eppn": "test"})
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

```json

```
