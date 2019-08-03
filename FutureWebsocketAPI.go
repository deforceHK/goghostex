package goghostex

type FutureWebsocketAPI interface {
	Init()

	Login(config *APIConfig) error

	Subscribe(topic []byte) error

	Unsubscribe(topic []byte) error

	Start()

	Close()
}
