package zk

import (
	"context"
	"errors"
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

func (c *Client) realPath(path string) string {
	if path == "" {
		return c.root
	}
	if path == "/" {
		return c.root
	}
	return c.root + path
}

// EnsurePath 确保路径存在，不存在则创建
func (c *Client) EnsurePath(path string) error {
	realPath := c.realPath(path)

	// 空路径或根路径，直接返回
	if realPath == "" || realPath == "/" {
		return nil
	}

	// 递归创建节点
	return c.ensurePathRecursive(realPath)
}

// ensurePathRecursive 递归创建节点
func (c *Client) ensurePathRecursive(path string) error {
	// 检查节点是否存在
	exists, stat, err := c.conn.Exists(path)
	if err != nil {
		return fmt.Errorf("检查节点 %s 是否存在失败: %w", path, err)
	}

	if exists {
		// 如果节点是临时节点，不能有子节点，返回错误
		if stat.EphemeralOwner != 0 {
			return fmt.Errorf("节点 %s 是临时节点，不能有子节点", path)
		}
		return nil
	}

	// 递归创建父节点
	parent := getParentPath(path)
	if parent != "" && parent != "/" {
		if err := c.ensurePathRecursive(parent); err != nil {
			return err
		}
	}

	// 创建当前节点
	_, err = c.conn.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
	if err != nil && !errors.Is(err, zk.ErrNodeExists) {
		return fmt.Errorf("创建节点 %s 失败: %w", path, err)
	}

	return nil
}
