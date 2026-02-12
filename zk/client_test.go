package zk

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testContainer testcontainers.Container
	testHost      string
	testPort      string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 启动 ZooKeeper 容器
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "zookeeper:3.6.4",
			ExposedPorts: []string{"2181/tcp"},
			Env: map[string]string{
				"ZOO_SERVER_MAX_CLIENT_CNXNS": "60",
				"ZOO_TICK_TIME":               "2000",
				"ALLOW_ANONYMOUS_LOGIN":       "yes",
			},
			WaitingFor: wait.ForLog("ZooKeeper JMX enabled by default"),
		},
		Started: true,
	})
	if err != nil {
		panic("启动 ZooKeeper 容器失败: " + err.Error())
	}

	testContainer = container

	// 获取容器映射的主机端口
	host, err := container.Host(ctx)
	if err != nil {
		panic("获取容器主机失败: " + err.Error())
	}
	testHost = host

	port, err := container.MappedPort(ctx, "2181")
	if err != nil {
		panic("获取映射端口失败: " + err.Error())
	}
	testPort = port.Port()
	log.Printf("启动 ZooKeeper 成功 %s:%s\n", testHost, testPort)

	// 运行测试
	exitCode := m.Run()

	// 清理容器
	container.Terminate(ctx)

	os.Exit(exitCode)
}

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	// 测试创建客户端
	tests := []struct {
		name           string
		addr           string
		sessionTimeout time.Duration
		wantRoot       string
		wantErr        bool
	}{
		{
			name:           "正常创建",
			addr:           testHost + ":" + testPort + "/tango",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/tango",
			wantErr:        false,
		},
		{
			name:           "带前缀",
			addr:           "zk://" + testHost + ":" + testPort + "/tango",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/tango",
			wantErr:        false,
		},
		{
			name:           "根路径为/",
			addr:           testHost + ":" + testPort + "/",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			client, err := NewClient(ctx, tt.addr, tt.sessionTimeout)
			if err != nil {
				if tt.wantErr {
					return
				}

				t.Errorf("NewClient() 意外错误: %v", err)
				return
			}

			defer client.Close()

			if tt.wantErr {
				t.Errorf("NewClient() 期望返回错误，但没有")
				return
			}

			if client.root != tt.wantRoot {
				t.Errorf("NewClient() root = %q, 期望 %q", client.root, tt.wantRoot)
			}
		})
	}
}

func TestNewClientTimeout(t *testing.T) {
	// 使用极短的超时上下文，不需要真实 ZK
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待超时生效，让 context 真正在使用前过期
	time.Sleep(50 * time.Millisecond)

	// 使用不存在的地址，context 应该先超时
	client, err := NewClient(ctx, "127.0.0.1:9999/tango", 5*time.Second)

	if err == nil {
		t.Error("NewClient() 期望返回超时错误，但没有")
		client.Close()
	}
}

func TestEnsurePath(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	tests := []struct {
		name      string
		root      string
		paths     []string
		wantError bool
	}{
		{
			name:      "创建单层路径",
			root:      "/test",
			paths:     []string{"/a"},
			wantError: false,
		},
		{
			name:      "创建多层路径",
			root:      "/test",
			paths:     []string{"/a/b/c"},
			wantError: false,
		},
		{
			name:      "创建重复路径",
			root:      "/test",
			paths:     []string{"/d", "/d"},
			wantError: false,
		},
		{
			name:      "创建空路径",
			root:      "/test",
			paths:     []string{""},
			wantError: false,
		},
		{
			name:      "创建根路径",
			root:      "/test",
			paths:     []string{"/"},
			wantError: false,
		},
		{
			name:      "创建多个相对路径",
			root:      "/test",
			paths:     []string{"/e/f", "/e/g", "/h"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			addr := testHost + ":" + testPort + tt.root
			client, err := NewClient(ctx, addr, 5*time.Second)
			if err != nil {
				t.Fatalf("创建客户端失败: %v", err)
			}
			defer client.Close()

			for _, path := range tt.paths {
				err := client.EnsurePath(path)
				if tt.wantError && err == nil {
					t.Errorf("EnsurePath(%q) 期望返回错误，但没有", path)
				} else if !tt.wantError && err != nil {
					t.Errorf("EnsurePath(%q) 意外错误: %v", path, err)
				}
			}

			// 验证路径确实存在
			for _, path := range tt.paths {
				if path == "" || path == "/" {
					continue
				}
				realPath := client.realPath(path)
				exists, _, err := client.conn.Exists(realPath)
				if err != nil {
					t.Errorf("验证节点 %s 存在失败: %v", path, err)
				} else if !exists {
					t.Errorf("节点 %s 应该存在但不存在", path)
				}
			}
		})
	}
}

func TestEnsurePathWithEphemeralNode(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := testHost + ":" + testPort + "/test-ephemeral"
	client, err := NewClient(ctx, addr, 5*time.Second)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 确保父节点路径存在
	if err := client.EnsurePath(""); err != nil {
		t.Fatalf("确保根路径存在失败: %v", err)
	}

	// 创建一个临时节点
	tempPath := "/ephemeral-node"
	realPath := client.realPath(tempPath)
	_, err = client.conn.Create(realPath, nil, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		t.Fatalf("创建临时节点失败: %v", err)
	}

	// 尝试在临时节点下创建子节点，应该报错
	childPath := tempPath + "/child"
	err = client.EnsurePath(childPath)
	if err == nil {
		t.Error("EnsurePath 应该在临时节点下创建子节点时返回错误，但没有")
	} else if !strings.Contains(err.Error(), "临时节点") {
		t.Errorf("EnsurePath 返回错误不正确: %v", err)
	}
}
