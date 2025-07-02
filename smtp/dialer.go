package smtp

import "time"

type Dialer struct {
	ProxyAddress string        // 代理地址，如 socks5://1.2.3.4:1080
	Timeout      time.Duration // 超时时间
}

