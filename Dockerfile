FROM golang:1.8

WORKDIR /go/src/github.com/guillaumerose/glcoud-tar
COPY . .

RUN go build

ENTRYPOINT ./glcoud-tar