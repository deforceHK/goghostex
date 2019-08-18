# goghostex


Goghostex is a open source API of TOP crypto currency exchanges. You can use it directly for your data collector or trading program.

## Feature

The list of goghost supported API.As below:


| |SPON|MARGIN|FUTURE|SWAP|
|:---|:---|:---|:---|:---
|OKEX|YES|NO|YES|NO|
|BINANCE|YES|NO|NO|NO


## Clone

```
git clone https://github.com/strengthening/goghostex.git
```

## Install 

```
go install
```


## TODO

- Support bitmex exchange.
- Support bitstamp exchange.
- Add `cli` features.
- Update the order feature
    - FOK(Fill Or Kill)
    - IOC(Immediate Or Cancel)
    - ROD(Rest Of Day)


## LICENSE

The project use the [New BSD License](./LICENSE)

## Credits

- [gorilla/websocket](https://github.com/gorilla/websocket)
    - A WebSocket implementation for Go.
- [nntaoli-project/GoEx](https://github.com/nntaoli-project/GoEx.git)
    - A Exchange REST and WebSocket API for Golang.