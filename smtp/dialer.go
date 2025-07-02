package smtp

import "time"

// Dialer 用于定义代理连接信息和超时设置
type Dialer struct {
	Timeout      time.Duration // SMTP 操作超时时间
	ProxyAddress string        // 代理地址（格式如：socks5://127.0.0.1:1080）
}

