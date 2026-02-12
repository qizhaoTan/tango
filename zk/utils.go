package zk

import (
	"fmt"
	"strings"
)

// 样例 zk://host1,host2/root 或 host1,host2/root
func parse(addr string) (hosts []string, root string, err error) {
	const prefix = "zk://"
	if strings.HasPrefix(addr, prefix) {
		addr = addr[len(prefix):]
	}

	pos := strings.LastIndex(addr, "/")
	if pos < 0 {
		err = fmt.Errorf("invalid zk addr: %s", addr)
		return
	}

	hosts = strings.Split(addr[:pos], ",")
	root = addr[pos:]
	return
}

// getParentPath 获取路径的父路径
func getParentPath(path string) string {
	// 空路径或根路径，返回空
	if path == "" || path == "/" {
		return ""
	}

	// 移除尾部的/
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	lastSlash := strings.LastIndex(path, "/")
	if lastSlash < 0 {
		return ""
	}
	if lastSlash == 0 {
		return "/"
	}

	return path[:lastSlash]
}
