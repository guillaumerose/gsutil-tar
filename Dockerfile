FROM golang:1.8

WORKDIR /go/src/github.com/guillaumerose/gcloud-tar
COPY . .

RUN go build
