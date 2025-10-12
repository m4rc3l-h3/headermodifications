package headermodifications

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/m4rc3l-h3/headermodifications/pkg/handler/deleter"
	"github.com/m4rc3l-h3/headermodifications/pkg/handler/join"
	"github.com/m4rc3l-h3/headermodifications/pkg/handler/rename"
	"github.com/m4rc3l-h3/headermodifications/pkg/handler/rewrite"
	"github.com/m4rc3l-h3/headermodifications/pkg/handler/set"
	"github.com/m4rc3l-h3/headermodifications/pkg/handler/set_first"
	"github.com/m4rc3l-h3/headermodifications/pkg/types"
)

// HeadersModifications holds the necessary components of a Traefik plugin.
type HeadersModifications struct {
	name         string
	next         http.Handler
	reqHandlers  []types.Handler
	respHandlers []types.Handler
}

// Config holds configuration to be passed to the plugin.
type Config struct {
	Rules []types.Rule
}

// CreateConfig populates the Config data object.
func CreateConfig() *Config {
	return &Config{
		Rules: []types.Rule{},
	}
}

// New instantiates and returns the required components used to handle an HTTP request.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	handlerBuilder := map[types.RuleType]func(types.Rule) (types.Handler, error){
		types.Delete:           deleter.New,
		types.Join:             join.New,
		types.Rename:           rename.New,
		types.RewriteValueRule: rewrite.New,
		types.Set:              set.New,
		types.SetFirst:         set_first.New,
	}

	reqHandlers := make([]types.Handler, 0, len(config.Rules))
	respHandlers := make([]types.Handler, 0, len(config.Rules))

	for _, rule := range config.Rules {
		newHandler, ok := handlerBuilder[rule.Type]
		if !ok {
			return nil, fmt.Errorf("%w: %s", types.ErrInvalidRuleType, rule.Name)
		}

		handler, err := newHandler(rule)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, rule.Name)
		}

		if err := handler.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %s", err, rule.Name)
		}

		if rule.SetOnResponse {
			respHandlers = append(respHandlers, handler)
		} else {
			reqHandlers = append(reqHandlers, handler)
		}
	}

	return &HeadersModifications{
		name:         name,
		next:         next,
		reqHandlers:  reqHandlers,
		respHandlers: respHandlers,
	}, nil
}

// Iterate over every header to match the ones specified in the config and
// return nothing if regexp failed.
func (u *HeadersModifications) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	for _, handler := range u.reqHandlers {
		handler.Handle(responseWriter, request)
	}

	wrappedResponseWriter := newWrappedResponseWriter(responseWriter, func(rw http.ResponseWriter) {
		for _, handler := range u.respHandlers {
			handler.Handle(rw, request)
		}
	})

	u.next.ServeHTTP(wrappedResponseWriter, request)
}

type wrappedResponseWriter struct {
	rw         http.ResponseWriter
	handler    func(http.ResponseWriter)
	headerSent bool
}

func newWrappedResponseWriter(rw http.ResponseWriter, handler func(http.ResponseWriter)) http.ResponseWriter {
	return &wrappedResponseWriter{
		rw:         rw,
		handler:    handler,
		headerSent: false,
	}
}

func (wrw *wrappedResponseWriter) handleResponseHeader() {
	if wrw.headerSent {
		return
	}

	wrw.headerSent = true
	wrw.handler(wrw.rw)
}

func (wrw *wrappedResponseWriter) Header() http.Header {
	return wrw.rw.Header()
}

func (wrw *wrappedResponseWriter) Write(p []byte) (int, error) {
	wrw.handleResponseHeader()

	n, err := wrw.rw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("write response: %w", err)
	}

	return n, nil
}

func (wrw *wrappedResponseWriter) WriteHeader(statusCode int) {
	wrw.handleResponseHeader()
	wrw.rw.WriteHeader(statusCode)
}

func (wrw *wrappedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := wrw.rw.(http.Hijacker)
	if !ok {
		return nil, nil, types.ErrNotHTTPHijacker
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, fmt.Errorf("hijack: %w", err)
	}

	return conn, rw, nil
}
