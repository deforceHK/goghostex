FROM golang:alpine
RUN mkdir -p /go/src/github.com/strengthening/goghostex
COPY . /go/src/github.com/strengthening/goghostex
