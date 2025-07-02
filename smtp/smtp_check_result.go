package smtp

// SMTPCheckResult 表示一次 SMTP 验证的结果
type SMTPCheckResult struct {
	Deliverable bool   // 是否可投递
	CatchAll    bool   // 是否为 catch-all 域名（可选，根据需要可用）
	Error       string // 如果失败，记录错误信息（可选）
}

