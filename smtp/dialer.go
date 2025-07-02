// 文件位置：smtp/dialer.go
// 用于支持 SOCKS5/HTTP 代理拨号连接 SMTP 服务器

package smtp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// Dialer 用于 SMTP 检测的拨号器，支持代理
type Dialer struct {
	Timeout      time.Duration
	ProxyAddress string // 例如 socks5://127.0.0.1:1080
}

// Dial 建立与 SMTP 服务器的连接
func (d *Dialer) Dial(ctx context.Context, network, address string) (net.Conn, error) {
	if d.ProxyAddress == "" {
		return net.DialTimeout(network, address, d.Timeout)
	}

	var auth proxy.Auth
	proxyAddr := d.ProxyAddress

	if strings.HasPrefix(proxyAddr, "socks5://") {
		proxyAddr = strings.TrimPrefix(proxyAddr, "socks5://")
	} else {
		return nil, fmt.Errorf("仅支持 socks5 代理: %s", d.ProxyAddress)
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, &auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("创建 SOCKS5 代理失败: %w", err)
	}

	conn, err := dialer.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("代理拨号失败: %w", err)
	}

	return conn, nil
}

