FROM golang:latest

ADD . /go/src/github.com/cj123/wordsearch

WORKDIR /go/src/github.com/cj123/wordsearch

RUN go get .
RUN go build .

EXPOSE 5598

ENTRYPOINT /go/src/github.com/cj123/wordsearch/wordsearch
