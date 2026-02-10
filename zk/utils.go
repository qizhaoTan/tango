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
