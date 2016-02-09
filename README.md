# Subjects Reader/Writer for Neo4j (subjects-rw-neo4j)
[![Circle CI](https://circleci.com/gh/Financial-Times/subjects-rw-neo4j/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/subjects-rw-neo4j/tree/master)

__An API for reading/writing subjects into Neo4j. Expects the subjects json supplied to be in the format that comes out of the subjects transformer.__

## Installation

For the first time:

`go get github.com/Financial-Times/subjects-rw-neo4j`

or update:

`go get -u github.com/Financial-Times/subjects-rw-neo4j`

## Running

```
export|set PORT=8080
export|set NEO_URL={neo4jUrl}
export|set BATCH_SIZE=50
export|set GRAPHITE_TCP_ADDRESS=graphite.ft.com:2003
export|set GRAPHITE_PREFIX=coco.{env}.services.subjects-rw-neo4j.{instanceNumber}
export|set LOG_METRICS=true
$GOPATH/bin/subjects-rw-neo4j
```
Flags are also supported

```
$GOPATH/bin/subjects-rw-neo4j --neo-url={neo4jUrl} --port=8080 --batch-size=50 --graphite-tcp-address=graphite.ft.com:2003 --graphite-prefix=coco.{env}.services.subjects-rw-neo4j.{instanceNumber} --log-metrics=true
```

With Docker:

`docker build -t coco/subjects-rw-neo4j .`

`docker run -ti --env NEO_URL=<base url> coco/subjects-rw-neo4j`

## Endpoints
/subjects/{uuid}
### PUT
The only mandatory field is the uuid, and the uuid in the body must match the one used on the path.
Every request results in an attempt to update that subject: unlike with GraphDB there is no check on whether the subject already exists and whether there are any changes between what's there and what's being written. We just do a MERGE which is Neo4j for create if not there, update if it is there.

Example:
`curl -XPUT -H "X-Request-Id: 123" -H "Content-Type: application/json" localhost:8080/subjects/bba39990-c78d-3629-ae83-808c333c6dbc --data '{"uuid":"bba39990-c78d-3629-ae83-808c333c6dbc","canonicalName":"Metals Markets","tmeIdentifier":"MTE3-U3ViamVjdHM=","type":"Subject"}'`

### GET
Will return 200 if successful, 404 if not found
`curl -H "X-Request-Id: 123" localhost:8080/subjects/bba39990-c78d-3629-ae83-808c333c6dbc`

### DELETE
Will return 204 if successful, 404 if not found
`curl -XDELETE -H "X-Request-Id: 123" localhost:8080/subjects/bba39990-c78d-3629-ae83-808c333c6dbc`
