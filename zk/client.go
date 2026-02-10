package zk

import (
	"time"
)

type Client struct {
}

func NewClient(addr string, sessionTimeout time.Duration) *Client {
	//conn, eventChan, err := zk.Connect(addr, sessionTimeout)
	//if err != nil {
	//	panic(err)
	//}
	//defer conn.Close()
	//
	//// 等待会话建立
	//for event := range eventChan {
	//	if event.State == zk.StateHasSession {
	//		break
	//	}
	//}

	return &Client{}
}
