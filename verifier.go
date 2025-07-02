package emailverifier

import (
	"github.com/sky668888/email-verifier/smtp"
)

type Verifier struct {
	smtpCheckEnabled bool
	smtpDialer       *smtp.Dialer
}

// 创建新的验证器实例
func NewVerifier() *Verifier {
	return &Verifier{}
}

// 启用 SMTP 检查（可自定义 dialer，包括代理支持）
func (v *Verifier) EnableSMTPCheckWithDialer(dialer smtp.Dialer) *Verifier {
	v.smtpCheckEnabled = true
	v.smtpDialer = &dialer
	return v
}

// 验证邮箱
func (v *Verifier) Verify(email string) (*Result, error) {
	result := &Result{Email: email}

	if v.smtpCheckEnabled && v.smtpDialer != nil {
		smtpResult, err := v.smtpDialer.Verify(email)
		if err != nil {
			return result, err
		}
		result.SMTP = smtpResult
	}

	return result, nil
}

// 结果结构体，包含 SMTP 检查结果
type Result struct {
	Email string
	SMTP  *smtp.SMTPCheckResult
}

