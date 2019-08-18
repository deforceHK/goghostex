# goghostex

[![CircleCI](https://circleci.com/gh/strengthening/goghostex.svg?style=svg)](https://circleci.com/gh/strengthening/goghostex)
[![CircleCI](https://img.shields.io/badge/license-BSD-blue)](https://img.shields.io/badge/license-BSD-blue)

README: [English](https://github.com/strengthening/goghostex/blob/master/README.md) | [中文](https://github.com/strengthening/goghostex/blob/master/README-zh.md)

Goghostex是一个开源的头部数字货币交易所API。您可以直接用来搜集数据和交易程序。

## 特性

goghostex支持的交易所API。如下：

| |现货|杠杠|交割合约|永续合约|
|:---|:---|:---|:---|:---
|OKEX|YES|NO|YES|NO|
|BINANCE|YES|NO|NO|NO




## Clone

```
git clone https://github.com/strengthening/goghostex.git
```

## 安装 

```
go install
```


## 待完成

- Support bitmex exchange.
- Support bitstamp exchange.
- Add `cli` features.
- Update the order feature
    - FOK(Fill Or Kill)
    - IOC(Immediate Or Cancel)
    - ROD(Rest Of Day)


## 协议

The project use the [New BSD License](./LICENSE)

## 鸣谢

- [gorilla/websocket](https://github.com/gorilla/websocket)
    - A WebSocket implementation for Go.
- [nntaoli-project/GoEx](https://github.com/nntaoli-project/GoEx.git)
    - A Exchange REST and WebSocket API for Golang.