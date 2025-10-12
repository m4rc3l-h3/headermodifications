package set_from_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomMoulard/htransformation/pkg/handler/set_from"
	"github.com/tomMoulard/htransformation/pkg/tests/assert"
	"github.com/tomMoulard/htransformation/pkg/tests/require"
	"github.com/tomMoulard/htransformation/pkg/types"
)

func TestSetFromHandler(t *testing.T) {
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
			name: "Set from existing header",
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
			name: "Set from existing header to new header",
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
			name: "Set from non-existing header",
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
			name: "Set from non-existing header to new header",
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/foo", nil)
			require.NoError(t, err)

			for hName, hVal := range test.requestHeaders {
				req.Header.Add(hName, hVal)
			}

			setHandler, err := set_from.New(test.rule)
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
				Value:        "X-From",
				HeaderPrefix: "^",
				Type:         types.SetFrom,
			},
			wantErr: true,
		},
		{
			name: "missing HeaderPrefix",
			rule: types.Rule{
				Header: "not-empty",
				Value:  "X-From",
				Type:   types.SetFrom,
			},
			wantErr: true,
		},
		{
			name: "missing Value",
			rule: types.Rule{
				Header:       "not-empty",
				HeaderPrefix: "^",
				Type:         types.SetFrom,
			},
			wantErr: true,
		},
		{
			name: "valid rule",
			rule: types.Rule{
				Header:       "not-empty",
				HeaderPrefix: "^",
				Value:        "X-From",
				Type:         types.SetFrom,
			},
			wantErr: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			setHandler, err := set_from.New(test.rule)
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
