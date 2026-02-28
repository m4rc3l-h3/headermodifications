package types

import (
	"errors"
	"net/http"
	"regexp"
)

// RuleType define the possible types of rules.
type RuleType string

const (
	// Set will set the value of a header.
	Set RuleType = "Set"
	// Join will concatenate the values of headers.
	Join RuleType = "Join"
	// Delete will delete the value of a header.
	Delete RuleType = "Del"
	// Rename will rename a header.
	Rename RuleType = "Rename"
	// RewriteValueRule will replace the value of a header with the provided value.
	RewriteValueRule RuleType = "RewriteValueRule"
	// SetFrist will set the value of a header to the first non-empty header value of anohter header
	SetFirst     RuleType = "SetFirst"
	AllowClients RuleType = "AllowClients"
)

// Rule struct so that we get traefik config.
type Rule struct {
	Header       string         `yaml:"Header"`       // header value
	HeaderPrefix string         `yaml:"HeaderPrefix"` // header prefix to find header
	Name         string         `yaml:"Name"`         // rule name
	Regexp       *regexp.Regexp `yaml:"-"`            // Used for rewrite, rename header matching
	Sep          string         `yaml:"Sep"`          // separator to use for join
	Type         RuleType       `yaml:"Type"`         // Differentiate rule types
	Value        string         `yaml:"Value"`
	ValueReplace string         `yaml:"ValueReplace"` // value used as replacement in rewrite
	Values       []string       `yaml:"Values"`       // values to join
	// if SetOnResponse is true, the header will be changed on the response. It will be on the request otherwise (default).
	SetOnResponse bool `yaml:"SetOnResponse"`
	Debug         bool `json:"debug,omitempty"`

	// Value to use for LAN requests, e.g. "lan"
	LanValue string `json:"lanValue,omitempty" yaml:"LanValue,omitempty"`
	// Value to use for WAN requests, e.g. "wan"
	WanValue string `json:"wanValue,omitempty" yaml:"WanValue,omitempty"`
	// e.g. "X-Forwarded-For" or "X-Real-IP"
	TrustedHeader string `json:"trustedHeader,omitempty" yaml:"TrustedHeader,omitempty"`
	// Whether LAN is allowed
	AllowLan bool `json:"allowLan,omitempty" yaml:"AllowLan,omitempty"`
	// Allowed WAN IPs/CIDRs, e.g. ["203.0.113.5/32"]
	WanIPs []string `json:"wanIPs,omitempty" yaml:"WanIPs,omitempty"`
	// HTTP status for rejects (default 403)
	RejectStatus int `json:"rejectStatus,omitempty" yaml:"RejectStatus,omitempty"`
}

var ErrMissingRequiredFields = errors.New("missing required fields")

var ErrInvalidRuleType = errors.New("invalid rule type")

var ErrInvalidRegexp = errors.New("invalid regexp")

var ErrNotHTTPHijacker = errors.New("not an http.Hijacker")

type Handler interface {
	Validate() error
	Handle(rw http.ResponseWriter, req *http.Request) (blocked bool)
	Debug() bool
}
