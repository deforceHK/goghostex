# goghostex

[![img.shields.io](https://img.shields.io/badge/license-BSD-blue)](https://img.shields.io/badge/license-BSD-blue)
[![CircleCI](https://circleci.com/gh/strengthening/goghostex/tree/master.svg?style=svg&circle-token=3e0fb98af6c242519e447954d79a2188ef1bafa6)](https://circleci.com/gh/strengthening/goghostex/tree/master)

README: [English](https://github.com/strengthening/goghostex/blob/master/README.md) | [中文](https://github.com/strengthening/goghostex/blob/master/README-zh.md)

Goghostex是一个开源的头部数字货币交易所API。您可以直接用来搜集数据和交易程序。

## 特性

goghostex支持的交易所API。如下：

| |现货|杠杠|交割合约|永续合约|
|:---|:---|:---|:---|:---
|OKEX|YES|YES|YES|NO|
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

- Add `cli` features.
- Support bitstamp exchange.
- Support bitmex exchange.


## 协议

The project use the [New BSD License](./LICENSE)

## 鸣谢

- [gorilla/websocket](https://github.com/gorilla/websocket)
    - A WebSocket implementation for Go.
- [nntaoli-project/GoEx](https://github.com/nntaoli-project/GoEx.git)
    - A Exchange REST and WebSocket API for Golang.