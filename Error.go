package goghostex

import "fmt"

type WSRestartError struct {
	//Code int
	Msg string
}

func (e *WSRestartError) Error() string {
	return fmt.Sprintf("websocket restart error: %s", e.Msg)
}

type WSStopError struct {
	//Code int
	Msg string
}

func (e *WSStopError) Error() string {
	return fmt.Sprintf("websocket stop error: %s", e.Msg)
}
