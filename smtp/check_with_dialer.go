package smtp

import (
	"errors"
	"net"
	"net/smtp"
	"net/url"
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

	// 构建连接
	var conn net.Conn
	if dialer.ProxyAddress != "" {
		// 解析 socks5 代理 URI，例如：socks5h://user:pass@host:port
		proxyURL, err := url.Parse(dialer.ProxyAddress)
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}

		auth := &proxy.Auth{}
		if proxyURL.User != nil {
			auth.User = proxyURL.User.Username()
			auth.Password, _ = proxyURL.User.Password()
		}

		address := proxyURL.Host // host:port

		dialerProxy, err := proxy.SOCKS5("tcp", address, auth, proxy.Direct)
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}

		conn, err = dialerProxy.Dial("tcp", mxHost+":25")
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}
	} else {
		// 无代理连接
		conn, err = net.DialTimeout("tcp", mxHost+":25", dialer.Timeout)
		if err != nil {
			return &SMTPCheckResult{Deliverable: false}, err
		}
	}

	defer conn.Close()

	// SMTP 交互
	client, err := smtp.NewClient(conn, mxHost)
	if err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}
	defer client.Close()

	host := "localhost"
	if err = client.Hello(host); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	if err = client.Mail("tester@" + domain); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	if err = client.Rcpt(username + "@" + domain); err != nil {
		return &SMTPCheckResult{Deliverable: false}, err
	}

	// 邮箱可投递
	return &SMTPCheckResult{Deliverable: true}, nil
}
