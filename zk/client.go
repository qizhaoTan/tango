package zk

import (
	"context"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

type Client struct {
	conn *zk.Conn // ZooKeeper 连接
	root string   // 根节点
}

func NewClient(ctx context.Context, addr string, sessionTimeout time.Duration) (*Client, error) {
	hosts, root, err := parse(addr)
	if err != nil {
		return nil, err
	}

	conn, eventChan, err := zk.Connect(hosts, sessionTimeout)
	if err != nil {
		return nil, err
	}

	// 等待会话建立或超时
	for {
		select {
		case event := <-eventChan:
			if event.State == zk.StateHasSession {
				return &Client{conn: conn, root: root}, nil
			}
		case <-ctx.Done():
			conn.Close()
			return nil, fmt.Errorf("连接超时: %w", ctx.Err())
		}
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
