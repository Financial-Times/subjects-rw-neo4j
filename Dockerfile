
FROM alpine:3.3

ADD *.go /subjects-rw-neo4j/
ADD subjects/*.go /subjects-rw-neo4j/subjects/

RUN apk add --update bash \
  && apk --update add git bzr \
  && apk --update add go \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/subjects-rw-neo4j" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && mv subjects-rw-neo4j/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get -t ./... \
  && go build \
  && mv subjects-rw-neo4j /app \
  && apk del go git bzr \
  && rm -rf $GOPATH /var/cache/apk/*

CMD [ "/app" ] 