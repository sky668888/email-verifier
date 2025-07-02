package smtp

import (
	"errors"
	"net"
	"net/smtp"
	"strings"

	"golang.org/x/net/proxy"
)

// CheckSMTPWithDialer 通过 SOCKS5 代理进行 SMTP 验证
func CheckSMTPWithDialer(domain, username string, dialer Dialer) (*SMTPCheckResult, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return &SMTPCheckResult{Deliverable: false}, errors.New("no MX records found for domain")
	}

	// 优先使用优先级高的 MX 服务器
	mxHost := strings.TrimSuffix(mxRecords[0].Host, ".")

	// 使用代理构建连接器
	var conn net.Conn
	if dialer.ProxyAddress != "" {
		dialSocksProxy, err := proxy.SOCKS5("tcp", dialer.ProxyAddress, nil, proxy.Direct)
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}

		conn, err = dialSocksProxy.Dial("tcp", mxHost+":25")
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}
	} else {
		// 无代理，使用标准连接
		conn, err = net.DialTimeout("tcp", mxHost+":25", dialer.Timeout)
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}
	}

	defer conn.Close()

	// 构建 SMTP 客户端
	client, err := smtp.NewClient(conn, mxHost)
	if err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}
	defer client.Close()

	// 发送 EHLO
	host := "localhost"
	if err = client.Hello(host); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	// 发送 MAIL FROM
	if err = client.Mail("tester@" + domain); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	// 发送 RCPT TO
	if err = client.Rcpt(username + "@" + domain); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	// 如果上面都通过，说明邮箱可投递
	return &SMTPCheckResult{Deliverable: true}, nil
}

