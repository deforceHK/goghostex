package goghostex

type FutureWebsocketAPI interface {
	GetExchangeName() string

	Login()

	Subscribe()

	Unsubscribe()

	ReceiveMessage(msg <-chan string)
}
