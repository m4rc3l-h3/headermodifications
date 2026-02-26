package set_first_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m4rc3l-h3/headermodifications/pkg/handler/set_first"
	"github.com/m4rc3l-h3/headermodifications/pkg/tests/assert"
	"github.com/m4rc3l-h3/headermodifications/pkg/tests/require"
	"github.com/m4rc3l-h3/headermodifications/pkg/tests/utils"
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
			name: "Set first both given",
			rule: types.Rule{
				Header: "X-Test",
				Values: []string{
					"^X-First",
					"^X-Second",
				},
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"Foo":      "Bar",
				"X-First":  "First",
				"X-Second": "Second",
			},
			wantOnRequest: map[string]string{
				"Foo":      "Bar",
				"X-First":  "First",
				"X-Second": "Second",
				"X-Test":   "First",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set first only second given",
			rule: types.Rule{
				Header: "X-Test",
				Values: []string{
					"^X-First",
					"^X-Second",
				},
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"Foo":      "Bar",
				"X-Second": "Second",
			},
			wantOnRequest: map[string]string{
				"Foo":      "Bar",
				"X-Second": "Second",
				"X-Test":   "Second",
			},
			expectedHost: "example.com",
		},
		{
			name: "Set first no headers given",
			rule: types.Rule{
				Header: "X-Test",
				Values: []string{
					"^X-First",
					"^X-Second",
				},
				HeaderPrefix: "^",
			},
			requestHeaders: map[string]string{
				"Foo": "Bar",
			},
			wantOnRequest: map[string]string{
				"Foo":    "Bar",
				"X-Test": "",
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

			setHandler, err := set_first.New(tc.rule)
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
			name: "missing Header value",
			rule: types.Rule{
				HeaderPrefix: "^",
				Values:       []string{"not-empty"},
				Type:         types.SetFirst,
			},
			wantErr: true,
		},
		{
			name: "missing HeaderPrefix",
			rule: types.Rule{
				Header: "not-empty",
				Values: []string{"not-empty"},
				Type:   types.SetFirst,
			},
			wantErr: true,
		},
		{
			name: "missing Values",
			rule: types.Rule{
				Header:       "not-empty",
				HeaderPrefix: "^",

				Type: types.SetFirst,
			},
			wantErr: true,
		},
		{
			name: "valid rule",
			rule: types.Rule{
				Header:       "not-empty",
				HeaderPrefix: "^",
				Values:       []string{"not-empty"},
				Type:         types.SetFirst,
			},
			wantErr: false,
		},
	}

	for _, test := range testCases {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			setHandler, err := set_first.New(tc.rule)
			require.NoError(t, err)

			err = setHandler.Validate()
			t.Log(err)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetHandlerWithDebug(t *testing.T) {

	rule := types.Rule{
		Name:         "debug-test",
		Header:       "X-Test",
		HeaderPrefix: "^",
		Values:       []string{"^X-First"},
		Debug:        true, // Enable debug
		Type:         types.SetFirst,
	}

	setHandler, err := set_first.New(rule)
	require.NoError(t, err)

	// Test debug logging with matching header
	logs := utils.CaptureLogs(t, func() {
		req, _ := http.NewRequestWithContext(t.Context(), "GET", "http://example.com", nil)
		req.Header.Set("X-First", "debug-value")
		setHandler.Handle(httptest.NewRecorder(), req)
	})

	assert.Contains(t, logs, "[DEBUG set_first]")
	assert.Contains(t, logs, "Request Host: example.com")
	assert.Contains(t, logs, "got headerValue='debug-value'")
	assert.Contains(t, logs, "MATCH!")
	assert.Contains(t, logs, "SUCCESS: Set debug-test='debug-value'")
}

func TestSetHandlerDebugNoMatch(t *testing.T) {

	rule := types.Rule{
		Name:         "debug-no-match",
		Header:       "X-Test",
		HeaderPrefix: "^",
		Values:       []string{"^X-Missing"},
		Debug:        true,
		Type:         types.SetFirst,
	}

	setHandler, err := set_first.New(rule)
	require.NoError(t, err)

	logs := utils.CaptureLogs(t, func() {
		req, _ := http.NewRequestWithContext(t.Context(), "GET", "http://example.com", nil)
		req.Header.Set("X-Other", "some-value") // Wrong header - no match
		setHandler.Handle(httptest.NewRecorder(), req)
	})

	assert.Contains(t, logs, "[DEBUG set_first]")
	assert.Contains(t, logs, "No matching value found from 1 attempts")
	assert.NotContains(t, logs, "SUCCESS") // Should not succeed
}

func TestSetHandlerDebugEmptyValues(t *testing.T) {

	rule := types.Rule{
		Name:         "debug-empty",
		Header:       "X-Test",
		HeaderPrefix: "^",
		Values:       []string{"^X-Empty1", "^X-Empty2"},
		Debug:        true,
		Type:         types.SetFirst,
	}

	setHandler, err := set_first.New(rule)
	require.NoError(t, err)

	logs := utils.CaptureLogs(t, func() {
		req, _ := http.NewRequestWithContext(t.Context(), "GET", "http://example.com", nil)
		// No matching headers present
		setHandler.Handle(httptest.NewRecorder(), req)
	})

	assert.Contains(t, logs, "Trying value[0] '^X-Empty1': got headerValue=''")
	assert.Contains(t, logs, "Trying value[1] '^X-Empty2': got headerValue=''")
	assert.Contains(t, logs, "No matching value found from 2 attempts")
}
