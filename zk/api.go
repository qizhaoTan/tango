package zk

type IZkClient interface {
}

type WatchHandle = func(IZkClient, []byte, ...interface{}) error
type WatchChildrenHandle = func(IZkClient, []string, ...interface{}) error
