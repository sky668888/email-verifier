package emailverifier

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sky668888/email-verifier/smtp"
)

var (
	ErrEmptyEmail         = errors.New("email is empty")
	ErrInvalidFormatEmail = errors.New("email format is invalid")
)

type Verifier struct {
	disposableCheckEnabled bool
	roleCheckEnabled       bool
	mxCheckEnabled         bool
	smtpCheckEnabled       bool
	gravatarCheckEnabled   bool
	timeout                time.Duration

	// ✅ 新增 smtpDialer 字段
	smtpDialer *smtp.Dialer
}

func NewVerifier() *Verifier {
	return &Verifier{
		timeout:     15 * time.Second,
		smtpDialer:  nil, // 默认 nil
	}
}

func (v *Verifier) EnableDisposableCheck() *Verifier {
	v.disposableCheckEnabled = true
	return v
}

func (v *Verifier) EnableRoleCheck() *Verifier {
	v.roleCheckEnabled = true
	return v
}

func (v *Verifier) EnableMxCheck() *Verifier {
	v.mxCheckEnabled = true
	return v
}

func (v *Verifier) EnableSMTPCheck() *Verifier {
	v.smtpCheckEnabled = true
	return v
}

// ✅ 新增方法：支持自定义 Dialer（例如 SOCKS5）
func (v *Verifier) EnableSMTPCheckWithDialer(d smtp.Dialer) *Verifier {
	v.smtpCheckEnabled = true
	v.smtpDialer = &d
	return v
}

func (v *Verifier) Verify(email string) (*Result, error) {
	if strings.TrimSpace(email) == "" {
		return nil, ErrEmptyEmail
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !isValidEmailFormat(email) {
		return nil, ErrInvalidFormatEmail
	}

	result := &Result{Email: email}

	if v.disposableCheckEnabled {
		result.Disposable = isDisposableEmail(email)
	}

	if v.roleCheckEnabled {
		result.Role = isRoleEmail(email)
	}

	if v.mxCheckEnabled {
		result.MxRecord = checkMX(email)
	}

	if v.smtpCheckEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
		defer cancel()

		var smtpRes *SMTPResult
		var err error

		if v.smtpDialer != nil {
			smtpRes, err = checkSMTPWithDialer(ctx, email, *v.smtpDialer)
		} else {
			smtpRes, err = checkSMTP(ctx, email)
		}

		result.SMTP = smtpRes
		if err != nil {
			return result, fmt.Errorf("smtp check error: %w", err)
		}
	}

	return result, nil
}

// 简单邮箱格式校验
func isValidEmailFormat(email string) bool {
	// 使用简单的正则校验
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

