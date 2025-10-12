module github.com/m4rc3l-h3/htransformation

go 1.24

toolchain go1.24.1

tool (
	github.com/client9/misspell/cmd/misspell
	github.com/golangci/golangci-lint/cmd/golangci-lint
	github.com/goreleaser/goreleaser/v2
	github.com/traefik/yaegi
	golang.org/x/vuln/cmd/govulncheck
)
