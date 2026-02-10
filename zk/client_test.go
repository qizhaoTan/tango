package zk

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

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
		t.Fatalf("启动 ZooKeeper 容器失败: %v", err)
	}
	defer container.Terminate(ctx)

	// 获取容器映射的主机端口
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("获取容器主机失败: %v", err)
	}
	port, err := container.MappedPort(ctx, "2181")
	if err != nil {
		t.Fatalf("获取映射端口失败: %v", err)
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
			addr:           host + ":" + port.Port() + "/tango",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/tango",
			wantErr:        false,
		},
		{
			name:           "带前缀",
			addr:           "zk://" + host + ":" + port.Port() + "/tango",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/tango",
			wantErr:        false,
		},
		{
			name:           "根路径为/",
			addr:           host + ":" + port.Port() + "/",
			sessionTimeout: 5 * time.Second,
			wantRoot:       "/",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.addr, tt.sessionTimeout)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() 期望返回错误，但没有")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() 意外错误: %v", err)
				return
			}

			if client.root != tt.wantRoot {
				t.Errorf("NewClient() root = %q, 期望 %q", client.root, tt.wantRoot)
			}
		})
	}
}
