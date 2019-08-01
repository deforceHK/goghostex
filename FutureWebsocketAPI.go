package goghostex

type FutureWebsocketAPI interface {
	GetExchangeName() string

	Login() // 开始时必须登录，除非设置的是不用登录的

	Ping(pingTime int64, pongTime int64)  // 判断ping不同的时间，如果超过多长时间发送ping，如果超过多长时间没有接受到pong， 这重启。

	Subscribe()

	Unsubscribe()

	ReceiveMessage(msg <-chan string)
}
