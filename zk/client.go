package zk

import (
	"time"

	"github.com/go-zookeeper/zk"
)

type Client struct {
	root string // 根节点
}

func NewClient(addr string, sessionTimeout time.Duration) (*Client, error) {
	hosts, root, err := parse(addr)
	if err != nil {
		return nil, err
	}

	conn, eventChan, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 等待会话建立
	for event := range eventChan {
		if event.State == zk.StateHasSession {
			break
		}
	}

	c := &Client{
		root: root,
	}
	return c, nil
}
