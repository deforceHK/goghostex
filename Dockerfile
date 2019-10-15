FROM registry.cn-hongkong.aliyuncs.com/strengthening/gobuild:latest
RUN mkdir -p /go/src/github.com/strengthening/goghostex
COPY . /go/src/github.com/strengthening/goghostex

