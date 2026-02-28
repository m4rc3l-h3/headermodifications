package allow_clients

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m4rc3l-h3/headermodifications/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowClientsHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rule           types.Rule
		requestHeaders map[string]string
		remoteAddr     string
		wantPass       bool
		wantHeader     string // Expected value of rule.Name header if passed
		wantStatus     int    // Expected response status
	}{
		// LAN tests - allowed when allowLan=true
		{
			name: "LAN IP via X-Forwarded-For - allowed",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				WanValue: "wan",
				AllowLan: true,
				WanIPs:   []string{},
				Debug:    true,
			},
			requestHeaders: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
			},
			remoteAddr: "10.0.0.1:1234",
			wantPass:   true,
			wantHeader: "lan",
			wantStatus: 200,
		},
		{
			name: "LAN IP via X-Real-IP - allowed",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				AllowLan: true,
			},
			requestHeaders: map[string]string{
				"X-Real-IP": "172.16.1.50",
			},
			remoteAddr: "172.16.1.50:1234",
			wantPass:   true,
			wantHeader: "lan",
			wantStatus: 200,
		},
		{
			name: "LAN link-local - allowed",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				AllowLan: true,
			},
			requestHeaders: map[string]string{
				"X-Forwarded-For": "169.254.1.1",
			},
			wantPass:   true,
			wantHeader: "lan",
			wantStatus: 200,
		},

		// WAN tests - requires matching wanIPs
		{
			name: "WAN IP matches whitelist - allowed",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				WanValue: "wan",
				WanIPs:   []string{"203.0.113.5/32", "198.51.100.10"},
			},
			requestHeaders: map[string]string{
				"X-Forwarded-For": "203.0.113.5, 198.51.100.1",
			},
			wantPass:   true,
			wantHeader: "wan",
			wantStatus: 200,
		},
		{
			name: "WAN IP CIDR match - allowed",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				WanValue: "wan",
				WanIPs:   []string{"198.51.100.0/24"},
			},
			requestHeaders: map[string]string{
				"X-Real-IP": "198.51.100.42",
			},
			wantPass:   true,
			wantHeader: "wan",
			wantStatus: 200,
		},
		{
			name: "WAN IP no match - rejected",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				WanValue: "wan",
				WanIPs:   []string{"203.0.113.5/32"},
			},
			requestHeaders: map[string]string{
				"X-Forwarded-For": "1.2.3.4, 5.6.7.8", // Doesn't match whitelist
			},
			wantPass:   false,
			wantStatus: 403,
		},

		// Edge cases
		{
			name: "No IPs available - rejected",
			rule: types.Rule{
				Name:   "X-Client-Origin",
				WanIPs: []string{"203.0.113.5"},
			},
			requestHeaders: map[string]string{},
			remoteAddr:     "[::]:1234", // Invalid IP
			wantPass:       false,
			wantStatus:     403,
		},
		{
			name: "LAN disabled - rejected",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				AllowLan: false,
				WanValue: "wan",
				WanIPs:   []string{"203.0.113.5"},
			},
			requestHeaders: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
			},
			wantPass:   false,
			wantStatus: 403,
		},
		{
			name: "Trusted header override - WAN",
			rule: types.Rule{
				Name:          "X-Client-Origin",
				WanValue:      "wan",
				TrustedHeader: "X-Custom-IP",
				WanIPs:        []string{"203.0.113.5"},
			},
			requestHeaders: map[string]string{
				"X-Custom-IP": "203.0.113.5",
			},
			wantPass:   true,
			wantHeader: "wan",
			wantStatus: 200,
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/foo", nil)
			require.NoError(t, err)

			// Set RemoteAddr
			req.RemoteAddr = tc.remoteAddr

			for hName, hVal := range tc.requestHeaders {
				req.Header.Add(hName, hVal)
			}

			handler, err := New(tc.rule)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.Handle(rw, req)

			// Check pass/fail
			if tc.wantPass {
				assert.Equal(t, 200, rw.Code, "should be 200")
				if tc.wantHeader != "" {
					assert.Equal(t, tc.wantHeader, req.Header.Get(tc.rule.Name))
				}
			} else {
				assert.Equal(t, tc.wantStatus, rw.Code, "should reject with correct status")
				assert.Contains(t, rw.Header().Get("X-Rejected-Reason"), "not", "should set reject reason")
			}
		})
	}
	t.Logf("All passed")
}

func TestValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		rule    types.Rule
		wantErr bool
	}{
		{
			name: "no name",
			rule: types.Rule{
				LanValue: "lan",
				AllowLan: true,
			},
			wantErr: true,
		},
		{
			name: "neither allowLan nor wanIPs",
			rule: types.Rule{
				Name: "X-Client-Origin",
			},
			wantErr: true,
		},
		{
			name: "valid - allowLan true",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				AllowLan: true,
			},
			wantErr: false,
		},
		{
			name: "valid - wanIPs configured",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				WanValue: "wan",
				WanIPs:   []string{"203.0.113.5/32"},
			},
			wantErr: false,
		},
		{
			name: "valid - both configured",
			rule: types.Rule{
				Name:     "X-Client-Origin",
				LanValue: "lan",
				WanValue: "wan",
				AllowLan: true,
				WanIPs:   []string{"203.0.113.5"},
			},
			wantErr: false,
		},
	}

	for _, test := range testCases {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, err := New(tc.rule)
			require.NoError(t, err)

			err = handler.Validate()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	t.Logf("All passed")
}

func TestIPMatching(t *testing.T) {
	t.Parallel()

	rule := types.Rule{
		WanIPs: []string{"203.0.113.0/24", "198.51.100.10/32", "10.0.0.5"},
		Debug:  true,
	}
	handler, _ := New(rule)

	tests := []struct {
		ip        string
		cidrs     []string
		wantMatch bool
	}{
		{"203.0.113.42", []string{"203.0.113.0/24"}, true},
		{"198.51.100.10", []string{"203.0.113.0/24"}, false},
		{"10.0.0.5", []string{"203.0.113.0/24"}, false},
		{"198.51.100.10", []string{"198.51.100.10/32"}, true},
		{"10.0.0.5", []string{"10.0.0.5"}, true},
		{"1.2.3.4", []string{"203.0.113.0/24", "198.51.100.10/32", "10.0.0.5"}, false},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			t.Logf("Testing IP %q against %v", test.ip, test.cidrs)
			match := handler.(*AllowClients).ipAllowed(test.ip, test.cidrs)
			t.Logf("Result: %v (expected %v)", match, test.wantMatch)
			assert.Equal(t, test.wantMatch, match)
		})
	}
}
