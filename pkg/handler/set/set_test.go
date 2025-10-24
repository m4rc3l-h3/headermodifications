package set_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m4rc3l-h3/headermodifications/pkg/handler/set"
	"github.com/m4rc3l-h3/headermodifications/pkg/tests/assert"
	"github.com/m4rc3l-h3/headermodifications/pkg/tests/require"
	"github.com/m4rc3l-h3/headermodifications/pkg/types"
)

func TestSetHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rule           types.Rule
		requestHeaders map[string]string
		wantOnRequest  map[string]string
		wantOnResponse map[string]string
		expectedHost   string
	}{
		{
			name: "Set one simple",
			rule: types.Rule{
				Header: "X-Test",
				Value:  "Tested",
			},
			requestHeaders: map[string]string{
				"Foo": "Bar",
			},
			wantOnRequest: map[string]string{
				"Foo":    "Bar",
				"X-Test": "Tested",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set already existing simple",
			rule: types.Rule{
				Header: "X-Test",
				Value:  "Tested",
			},
			requestHeaders: map[string]string{
				"Foo":    "Bar",
				"X-Test": "Bar",
			},
			wantOnRequest: map[string]string{
				"Foo":    "Bar",
				"X-Test": "Tested", // Override
			},
			expectedHost: "example.com",
		},
		{
			name: "Set on response",
			rule: types.Rule{
				Header:        "X-Test",
				Value:         "Tested",
				SetOnResponse: true,
			},
			requestHeaders: map[string]string{
				"Foo": "Bar",
			},
			wantOnRequest: map[string]string{
				"Foo": "Bar",
			},
			wantOnResponse: map[string]string{
				"X-Test": "Tested",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set Host header",
			rule: types.Rule{
				Header: "Host",
				Value:  "example.org",
			},
			expectedHost: "example.org",
		},
		{
			name: "Set overwrite existing header by another header value",
			rule: types.Rule{
				Header:       "X-Test",
				Value:        "^X-From",
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"X-Test": "Foo",
				"X-From": "Bar",
			},
			wantOnRequest: map[string]string{
				"X-Test": "Bar",
				"X-From": "Bar",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set new header using another header value",
			rule: types.Rule{
				Header:       "X-Test",
				Value:        "^X-From",
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"X-From": "Bar",
			},
			wantOnRequest: map[string]string{
				"X-Test": "Bar",
				"X-From": "Bar",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set overwrite header by non-existing referenced header",
			rule: types.Rule{
				Header:       "X-Test",
				Value:        "^X-From",
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"X-Test": "Foo",
			},
			wantOnRequest: map[string]string{
				"X-Test": "",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set create new header from referenced non-existing header",
			rule: types.Rule{
				Header:       "X-Test",
				Value:        "^X-From",
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{},
			wantOnRequest: map[string]string{
				"X-Test": "",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set create new header from prefix without value",
			rule: types.Rule{
				Header:       "X-Test",
				Value:        "^",
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{},
			wantOnRequest: map[string]string{
				"X-Test": "^",
			},
			expectedHost: "example.com",
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/foo", nil)
			require.NoError(t, err)

			for hName, hVal := range tc.requestHeaders {
				req.Header.Add(hName, hVal)
			}

			setHandler, err := set.New(tc.rule)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			setHandler.Handle(rw, req)

			for hName, hVal := range tc.wantOnRequest {
				assert.Equal(t, hVal, req.Header.Get(hName))
			}

			for hName, hVal := range tc.wantOnResponse {
				assert.Equal(t, hVal, rw.Header().Get(hName))
			}

			assert.Equal(t, tc.expectedHost, req.Host)
			assert.Equal(t, "example.com", req.URL.Host)
		})
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		rule    types.Rule
		wantErr bool
	}{
		{
			name:    "no rules",
			wantErr: true,
		},
		{
			name: "missing Header and value",
			rule: types.Rule{
				Type: types.Set,
			},
			wantErr: true,
		},
		{
			name: "missing value",
			rule: types.Rule{
				Header: "not-empty",
				Type:   types.Set,
			},
			wantErr: true,
		},
		{
			name: "missing header",
			rule: types.Rule{
				Value: "not-empty",
				Type:  types.Set,
			},
			wantErr: true,
		},
		{
			name: "valid header replacement rule without header prefix",
			rule: types.Rule{
				Header: "not-empty",
				Value:  "not-empty",
				Type:   types.Set,
			},
			wantErr: false,
		},
		{
			name: "valid header replacement rule with header prefix",
			rule: types.Rule{
				Header:       "not-empty",
				HeaderPrefix: "not-empty",
				Value:        "not-empty",
				Type:         types.Set,
			},
			wantErr: false,
		},
	}

	for _, test := range testCases {

		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setHandler, err := set.New(tc.rule)
			require.NoError(t, err)

			t.Logf("Test case: %s\nHeader: %q\nValue: %q\nError: %#v\nWantErr: %v", tc.name, tc.rule.Header, tc.rule.Value, err, tc.wantErr)

			err = setHandler.Validate()
			t.Log(err)

			if tc.wantErr {
				assert.Error(t, err)
				t.Log(err)

			} else {
				assert.NoError(t, err)
			}
		})
	}
}
