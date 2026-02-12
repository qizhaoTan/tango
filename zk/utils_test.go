package zk

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantHosts []string
		wantRoot  string
		wantErr   bool
	}{
		{
			name:      "带zk://前缀",
			addr:      "zk://127.0.0.1:2181/tango",
			wantHosts: []string{"127.0.0.1:2181"},
			wantRoot:  "/tango",
			wantErr:   false,
		},
		{
			name:      "不带前缀",
			addr:      "127.0.0.1:2181/tango",
			wantHosts: []string{"127.0.0.1:2181"},
			wantRoot:  "/tango",
			wantErr:   false,
		},
		{
			name:      "多个主机",
			addr:      "host1:2181,host2:2181,host3:2181/tango",
			wantHosts: []string{"host1:2181", "host2:2181", "host3:2181"},
			wantRoot:  "/tango",
			wantErr:   false,
		},
		{
			name:      "多个主机带前缀",
			addr:      "zk://host1:2181,host2:2181/tango",
			wantHosts: []string{"host1:2181", "host2:2181"},
			wantRoot:  "/tango",
			wantErr:   false,
		},
		{
			name:      "无效地址-缺少/",
			addr:      "127.0.0.1:2181",
			wantHosts: nil,
			wantRoot:  "",
			wantErr:   true,
		},
		{
			name:      "无效地址-空地址",
			addr:      "",
			wantHosts: nil,
			wantRoot:  "",
			wantErr:   true,
		},
		{
			name:      "根路径为/",
			addr:      "127.0.0.1:2181/",
			wantHosts: []string{"127.0.0.1:2181"},
			wantRoot:  "/",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hosts, root, err := parse(tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parse(%q) 期望返回错误，但没有", tt.addr)
				}
				return
			}

			if err != nil {
				t.Errorf("parse(%q) 意外错误: %v", tt.addr, err)
				return
			}

			if !equalStringSlice(hosts, tt.wantHosts) {
				t.Errorf("parse(%q) hosts = %v, 期望 %v", tt.addr, hosts, tt.wantHosts)
			}

			if root != tt.wantRoot {
				t.Errorf("parse(%q) root = %q, 期望 %q", tt.addr, root, tt.wantRoot)
			}
		})
	}
}

// equalStringSlice 比较两个字符串切片是否相等
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetParentPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		want     string
	}{
		{
			name: "单层路径",
			path: "/a",
			want: "/",
		},
		{
			name: "多层路径",
			path: "/a/b/c",
			want: "/a/b",
		},
		{
			name: "根路径",
			path: "/",
			want: "",
		},
		{
			name: "空路径",
			path: "",
			want: "",
		},
		{
			name: "带尾部斜杠的路径",
			path: "/a/b/",
			want: "/a",
		},
		{
			name: "两层路径",
			path: "/a/b",
			want: "/a",
		},
		{
			name: "斜杠结尾的根路径",
			path: "//",
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParentPath(tt.path)
			if got != tt.want {
				t.Errorf("getParentPath(%q) = %q, 期望 %q", tt.path, got, tt.want)
			}
		})
	}
}
