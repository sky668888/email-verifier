package smtp

type SMTPCheckResult struct {
	Deliverable bool
	// 可以根据需要添加更多字段，如 SMTP code、错误信息等
}

