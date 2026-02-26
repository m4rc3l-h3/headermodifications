package utils

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/m4rc3l-h3/headermodifications/pkg/tests/require"
)

// captureLogs captures log output during f() execution
func CaptureLogs(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origOut := log.Writer()
	log.SetOutput(w)
	defer func() {
		log.SetOutput(origOut)
		w.Close()
		r.Close()
	}()

	f()

	w.Close()
	outputBytes, err := io.ReadAll(r)
	require.NoError(t, err)

	return strings.TrimSpace(string(outputBytes))
}
