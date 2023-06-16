FROM golang:alpine
RUN mkdir -p /go/src/github.com/deforceHK/goghostex
COPY . /go/src/github.com/deforceHK/goghostex
