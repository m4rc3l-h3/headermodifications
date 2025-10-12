package set_first_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomMoulard/htransformation/pkg/handler/set_first"
	"github.com/tomMoulard/htransformation/pkg/tests/assert"
	"github.com/tomMoulard/htransformation/pkg/tests/require"
	"github.com/tomMoulard/htransformation/pkg/types"
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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/foo", nil)
			require.NoError(t, err)

			for hName, hVal := range test.requestHeaders {
				req.Header.Add(hName, hVal)
			}

			setHandler, err := set_first.New(test.rule)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			setHandler.Handle(rw, req)

			for hName, hVal := range test.wantOnRequest {
				assert.Equal(t, hVal, req.Header.Get(hName))
			}

			for hName, hVal := range test.wantOnResponse {
				assert.Equal(t, hVal, rw.Header().Get(hName))
			}

			assert.Equal(t, test.expectedHost, req.Host)
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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			setHandler, err := set_first.New(test.rule)
			require.NoError(t, err)

			err = setHandler.Validate()
			t.Log(err)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
