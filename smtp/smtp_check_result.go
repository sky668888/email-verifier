package smtp

type SMTP struct {
	Host        string `json:"host"`         // SMTP 主机名
	Port        int    `json:"port"`         // SMTP 端口
	Connectable bool   `json:"connectable"`  // 是否能连接
	Deliverable bool   `json:"deliverable"`  // 是否能投递
	CatchAll    bool   `json:"catch_all"`    // 是否为 catch-all 域
	Disabled    bool   `json:"disabled"`     // 是否禁用
	FullInbox   bool   `json:"full_inbox"`   // 是否邮箱已满
	Error       string `json:"error"`        // 错误信息
}

