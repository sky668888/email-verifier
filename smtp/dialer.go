package smtp

import (
	"errors"
	"strings"
	"time"
)

type Dialer struct {
	Timeout      time.Duration
	ProxyAddress string // socks5://127.0.0.1:1080
}

func (d *Dialer) Verify(email string) (*SMTPCheckResult, error) {
	// 模拟检测逻辑
	if strings.HasSuffix(email, "@example.com") {
		return &SMTPCheckResult{Deliverable: true}, nil
	}

	if strings.Contains(email, "invalid") {
		return &SMTPCheckResult{Deliverable: false}, errors.New("SMTP: 邮箱不可达")
	}

	// 默认模拟返回可达
	return &SMTPCheckResult{Deliverable: true}, nil
}

