package emailverifier

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sky668888/email-verifier/smtp"
)

// Verifier is an email verifier. Create one by calling NewVerifier
type Verifier struct {
	smtpCheckEnabled     bool                       // SMTP check enabled or disabled (disabled by default)
	catchAllCheckEnabled bool                       // SMTP catchAll check enabled or disabled (enabled by default)
	domainSuggestEnabled bool                       // whether suggest a most similar correct domain or not (disabled by default)
	gravatarCheckEnabled bool                       // gravatar check enabled or disabled (disabled by default)
	fromEmail            string                     // name to use in the `EHLO:` SMTP command, defaults to "user@example.org"
	helloName            string                     // email to use in the `MAIL FROM:` SMTP command. defaults to `localhost`
	schedule             *schedule                  // schedule represents a job schedule
	proxyURI             string                     // use a SOCKS5 proxy to verify the email,
	apiVerifiers         map[string]smtpAPIVerifier // currently support gmail & yahoo, further contributions are welcomed.

	connectTimeout   time.Duration // Timeout for establishing connections
	operationTimeout time.Duration // Timeout for SMTP operations (e.g., EHLO, MAIL FROM, etc.)
	smtpDialer       *smtp.Dialer   // 自定义 Dialer，支持 SOCKS5 代理
}

// Result is the result of Email Verification
type Result struct {
	Email        string    `json:"email"`
	Reachable    string    `json:"reachable"`
	Syntax       Syntax    `json:"syntax"`
	SMTP         *SMTP     `json:"smtp"`
	Gravatar     *Gravatar `json:"gravatar"`
	Suggestion   string    `json:"suggestion"`
	Disposable   bool      `json:"disposable"`
	RoleAccount  bool      `json:"role_account"`
	Free         bool      `json:"free"`
	HasMxRecords bool      `json:"has_mx_records"`
}

var additionalDisposableDomains map[string]bool = map[string]bool{}

func init() {
	for d := range disposableDomains {
		disposableSyncDomains.Store(d, struct{}{})
	}
}

func NewVerifier() *Verifier {
	return &Verifier{
		fromEmail:            defaultFromEmail,
		helloName:            defaultHelloName,
		catchAllCheckEnabled: true,
		apiVerifiers:         map[string]smtpAPIVerifier{},
		connectTimeout:       10 * time.Second,
		operationTimeout:     10 * time.Second,
	}
}

func (v *Verifier) Verify(email string) (*Result, error) {
	ret := Result{
		Email:     email,
		Reachable: reachableUnknown,
	}

	syntax := v.ParseAddress(email)
	ret.Syntax = syntax
	if !syntax.Valid {
		return &ret, nil
	}

	ret.Free = v.IsFreeDomain(syntax.Domain)
	ret.RoleAccount = v.IsRoleAccount(syntax.Username)
	ret.Disposable = v.IsDisposable(syntax.Domain)

	if ret.Disposable {
		return &ret, nil
	}

	mx, err := v.CheckMX(syntax.Domain)
	if err != nil {
		return &ret, err
	}
	ret.HasMxRecords = mx.HasMXRecord

	var smtpResult *SMTP
	if v.smtpDialer != nil {
		smtpResult, err = v.CheckSMTPWithDialer(syntax.Domain, syntax.Username, v.smtpDialer)
	} else {
		smtpResult, err = v.CheckSMTP(syntax.Domain, syntax.Username)
	}
	if err != nil {
		return &ret, err
	}
	ret.SMTP = smtpResult
	ret.Reachable = v.calculateReachable(smtpResult)

	if v.gravatarCheckEnabled {
		gravatar, err := v.CheckGravatar(email)
		if err != nil {
			return &ret, err
		}
		ret.Gravatar = gravatar
	}

	if v.domainSuggestEnabled {
		ret.Suggestion = v.SuggestDomain(syntax.Domain)
	}

	return &ret, nil
}

func (v *Verifier) AddDisposableDomains(domains []string) *Verifier {
	for _, d := range domains {
		additionalDisposableDomains[d] = true
		disposableSyncDomains.Store(d, struct{}{})
	}
	return v
}

func (v *Verifier) EnableGravatarCheck() *Verifier {
	v.gravatarCheckEnabled = true
	return v
}

func (v *Verifier) DisableGravatarCheck() *Verifier {
	v.gravatarCheckEnabled = false
	return v
}

func (v *Verifier) EnableSMTPCheck() *Verifier {
	v.smtpCheckEnabled = true
	return v
}

func (v *Verifier) EnableSMTPCheckWithDialer(d smtp.Dialer) *Verifier {
	v.smtpCheckEnabled = true
	v.smtpDialer = &d
	return v
}

func (v *Verifier) DisableSMTPCheck() *Verifier {
	v.smtpCheckEnabled = false
	return v
}

func (v *Verifier) EnableCatchAllCheck() *Verifier {
	v.catchAllCheckEnabled = true
	return v
}

func (v *Verifier) DisableCatchAllCheck() *Verifier {
	v.catchAllCheckEnabled = false
	return v
}

func (v *Verifier) EnableDomainSuggest() *Verifier {
	v.domainSuggestEnabled = true
	return v
}

func (v *Verifier) DisableDomainSuggest() *Verifier {
	v.domainSuggestEnabled = false
	return v
}

func (v *Verifier) EnableAutoUpdateDisposable() *Verifier {
	v.stopCurrentSchedule()
	_ = updateDisposableDomains(disposableDataURL)
	v.schedule = newSchedule(24*time.Hour, updateDisposableDomains, disposableDataURL)
	v.schedule.start()
	return v
}

func (v *Verifier) DisableAutoUpdateDisposable() *Verifier {
	v.stopCurrentSchedule()
	return v
}

func (v *Verifier) FromEmail(email string) *Verifier {
	v.fromEmail = email
	return v
}

func (v *Verifier) HelloName(domain string) *Verifier {
	v.helloName = domain
	return v
}

func (v *Verifier) Proxy(proxyURI string) *Verifier {
	v.proxyURI = proxyURI
	return v
}

func (v *Verifier) ConnectTimeout(timeout time.Duration) *Verifier {
	v.connectTimeout = timeout
	return v
}

func (v *Verifier) OperationTimeout(timeout time.Duration) *Verifier {
	v.operationTimeout = timeout
	return v
}

func (v *Verifier) EnableAPIVerifier(name string) error {
	switch name {
	case YAHOO:
		v.apiVerifiers[YAHOO] = newYahooAPIVerifier(http.DefaultClient)
	default:
		return fmt.Errorf("unsupported to enable the API verifier for vendor: %s", name)
	}
	return nil
}

func (v *Verifier) DisableAPIVerifier(name string) {
	delete(v.apiVerifiers, name)
}

func (v *Verifier) calculateReachable(s *SMTP) string {
	if !v.smtpCheckEnabled {
		return reachableUnknown
	}
	if s.Deliverable {
		return reachableYes
	}
	if s.CatchAll {
		return reachableUnknown
	}
	return reachableNo
}

func (v *Verifier) stopCurrentSchedule() {
	if v.schedule != nil {
		v.schedule.stop()
	}
}

