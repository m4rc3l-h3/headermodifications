package header_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m4rc3l-h3/headermodifications/pkg/tests/assert"
	"github.com/m4rc3l-h3/headermodifications/pkg/utils/header"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		header          string
		expectedHeaders map[string][]string
		expectedHost    string
	}{
		{
			name:   "Delete header",
			header: "Foo",
			expectedHeaders: map[string][]string{
				"Foo": {""},
			},
			expectedHost: "example.com",
		},
		{
			name:         "Delete Host header",
			header:       "Host",
			expectedHost: "",
			expectedHeaders: map[string][]string{
				"Foo": {"Bar"},
			},
		},
		{
			name:         "Delete empty header",
			header:       "",
			expectedHost: "example.com",
			expectedHeaders: map[string][]string{
				"Foo": {"Bar"},
			},
		},
		{
			name:         "Delete header with empty value",
			header:       "Foo",
			expectedHost: "example.com",
			expectedHeaders: map[string][]string{
				"Foo": {"Bar"},
			},
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
			req.Header.Set("Foo", "Bar")

			header.Delete(req, tc.header)

			assert.Equal(t, tc.expectedHost, req.Host)

			for hName, hVal := range req.Header {
				assert.Equalf(t, tc.expectedHeaders[hName], hVal, "header %q", hName)
			}
		})
	}
}
