package smtp

import "time"

type Dialer struct {
	Timeout      time.Duration
	ProxyAddress string
}

func CheckWithDialer(domain, username, helloName, fromEmail string, dialer Dialer) (*SMTP, error) {
	// 你原有的 SMTP 检测逻辑，支持代理 SOCKS5 的 dial 实现
	// 这里只是示意返回
	return &SMTP{Deliverable: true}, nil
}

