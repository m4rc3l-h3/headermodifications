package allow_clients

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/m4rc3l-h3/headermodifications/pkg/types"
)

type AllowClients struct {
	rule *types.Rule
}

func New(rule types.Rule) (types.Handler, error) {
	if rule.RejectStatus == 0 {
		rule.RejectStatus = 403 // Default reject = 403
	}
	return &AllowClients{rule: &rule}, nil
}

func (c *AllowClients) Debug() bool {
	return c.rule.Debug
}

func (c *AllowClients) Validate() error {
	if c.rule.Name == "" {
		return types.ErrMissingRequiredFields
	}
	if !c.rule.AllowLan && len(c.rule.WanIPs) == 0 {
		return types.ErrMissingRequiredFields
	}
	return nil
}

func (c *AllowClients) Handle(rw http.ResponseWriter, req *http.Request) (blocked bool) {
	if c.rule.Debug {
		fmt.Printf("[DEBUG AllowClients] Rule: %+v, Host: %s", c.rule, req.Host)
	}

	// 1. Log ALL available IP sources for debugging
	c.logAllIPSources(req)

	// 2. Determine client IP
	clientIP := c.getClientIP(req)
	if c.rule.Debug {
		fmt.Printf("[DEBUG AllowClients] Client IP=%s", clientIP)
	}
	if clientIP == "" {
		rejectRequest(rw, "No client IP found")
		return true
	}

	isLan := c.isPrivateIP(clientIP)
	// 3. Apply whitelist logic
	if isLan {
		if !c.rule.AllowLan {
			if c.rule.Debug {
				fmt.Printf("[DEBUG AllowClients] Client IP %s isLan=%t, allowLan=False", clientIP, isLan)
			}

			rejectRequest(rw, "LAN traffic not allowed")
			return true
		}
	} else {
		if c.rule.Debug {
			fmt.Printf("[DEBUG AllowClients] Client IP %s isLan=%t, allowLan=False in WAN check", clientIP, isLan)
		}
		// WAN traffic - check against whitelist
		if !c.ipAllowed(clientIP, c.rule.WanIPs) {
			if c.rule.Debug {
				fmt.Printf("[DEBUG AllowClients] Client IP %s isLan=%t, allowLan=False and WAN not matched", clientIP, isLan)
			}
			rejectRequest(rw, "WAN IP not whitelisted")
			return true
		}
	}

	// 4. Set header based on origin
	var value string
	if isLan {
		value = c.rule.LanValue
	} else {
		value = c.rule.WanValue
	}

	if value != "" {
		req.Header.Set(c.rule.Name, value)
		if c.rule.Debug {
			fmt.Printf("[DEBUG AllowClients] Set %s=%q for IP %s", c.rule.Name, value, clientIP)
		}
	}

	rw.WriteHeader(200) // "Authorization successful, continue"
	req.Header.Set(c.rule.Name, value)
	return false
}

// Log all possible IP sources for debugging
func (c *AllowClients) logAllIPSources(req *http.Request) {
	if !c.rule.Debug {
		return
	}

	fmt.Printf("[DEBUG AllowClients IPs] RemoteAddr=%q", req.RemoteAddr)
	fmt.Printf("[DEBUG AllowClients IPs] X-Forwarded-For=%q", req.Header.Get("X-Forwarded-For"))
	fmt.Printf("[DEBUG AllowClients IPs] X-Real-IP=%q", req.Header.Get("X-Real-IP"))
	fmt.Printf("[DEBUG AllowClients IPs] X-Forwarded-Host=%q", req.Header.Get("X-Forwarded-Host"))
	fmt.Printf("[DEBUG AllowClients IPs] X-Cluster-Client-IP=%q", req.Header.Get("X-Cluster-Client-IP"))
	fmt.Printf("[DEBUG AllowClients IPs] X-Original-Forwarded-For=%q", req.Header.Get("X-Original-Forwarded-For"))

	// Log all headers containing "ip" or "forward"
	for h := range req.Header {
		hLower := strings.ToLower(h)
		if strings.Contains(hLower, "ip") || strings.Contains(hLower, "forward") {
			fmt.Printf("[DEBUG AllowClients IPs] %s=%q", h, req.Header.Get(h))
		}
	}
}

func (c *AllowClients) getClientIP(req *http.Request) string {
	// 1) Configured trusted header
	if h := strings.TrimSpace(c.rule.TrustedHeader); h != "" {
		if v := req.Header.Get(h); v != "" {
			ip := firstForwardedIP(v)
			if ip != "" && net.ParseIP(ip) != nil {
				if c.rule.Debug {
					fmt.Printf("[DEBUG AllowClients] Using trusted header %q -> %q", h, ip)
				}
				return ip
			}
		}
	}

	// 2) X-Forwarded-For (first IP)
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ip := firstForwardedIP(xff)
		if ip != "" && net.ParseIP(ip) != nil {
			if c.rule.Debug {
				fmt.Printf("[DEBUG AllowClients] Using X-Forwarded-For -> %q", ip)
			}
			return ip
		}
	}

	// 3) X-Real-IP
	if xrip := req.Header.Get("X-Real-IP"); xrip != "" {
		ip := strings.TrimSpace(xrip)
		if net.ParseIP(ip) != nil {
			if c.rule.Debug {
				fmt.Printf("[DEBUG AllowClients] Using X-Real-IP -> %q", ip)
			}
			return ip
		}
	}

	// 4) RemoteAddr (last resort)
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		if c.rule.Debug {
			fmt.Printf("[DEBUG AllowClients] Using RemoteAddr -> %q", host)
		}
		return host
	}

	return ""
}

// Check if IP matches any allowed WAN IPs/CIDRs
func (c *AllowClients) ipAllowed(ipStr string, allowed []string) bool {

	fmt.Printf("[DEBUG AllowClients] Check ipAllowed for IP %s", ipStr)

	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("[DEBUG AllowClients] IP %s is not a valid IP address", ipStr)
		return false
	}

	fmt.Printf("[DEBUG AllowClients] Check IP is in %s", allowed)

	for _, allowedCIDR := range allowed {

		fmt.Printf("[DEBUG AllowClients] Check whether IP %s is in %q", ipStr, allowedCIDR)

		if c.ipMatchesCIDR(ipStr, allowedCIDR) {
			if c.rule.Debug {
				fmt.Printf("[DEBUG AllowClients] IP %s matches allowed %q", ipStr, allowedCIDR)
			}
			return true
		}
	}
	return false
}

// Simple CIDR matching (IPv4 only for simplicity)
func (c *AllowClients) ipMatchesCIDR(ipStr, cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Single IP
		return strings.EqualFold(ipStr, cidr)
	}
	ip := net.ParseIP(ipStr)
	return ipNet.Contains(ip)
}

// Helper functions
func firstForwardedIP(raw string) string {
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func (c *AllowClients) isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 || // 10.0.0.0/8
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16-31.0.0/12
			(ip4[0] == 192 && ip4[1] == 168) || // 192.168.0.0/16
			(ip4[0] == 169 && ip4[1] == 254) { // 169.254.0.0/16 link-local
			return true
		}
	}
	// IPv6 ULA fc00::/7
	if strings.HasPrefix(ipStr, "fc") || strings.HasPrefix(ipStr, "fd") {
		return true
	}
	return false
}

func rejectRequest(rw http.ResponseWriter, reason string) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("X-Rejected-Reason", reason)
	rw.WriteHeader(403)
	rw.Write([]byte("Access denied: " + reason))
}
