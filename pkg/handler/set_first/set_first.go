package set_first

import (
	"log"
	"net/http"
	"strings"

	"github.com/m4rc3l-h3/headermodifications/pkg/types"
	"github.com/m4rc3l-h3/headermodifications/pkg/utils"
)

type SetFirst struct {
	rule *types.Rule
}

func (s *SetFirst) Debug() bool {
	return s.rule.Debug
}

func New(rule types.Rule) (types.Handler, error) {
	return &SetFirst{rule: &rule}, nil
}

func (s *SetFirst) Validate() error {
	if s.rule.Header == "" || s.rule.HeaderPrefix == "" || len(s.rule.Values) == 0 {
		return types.ErrMissingRequiredFields
	}

	return nil
}

func (s *SetFirst) Handle(rw http.ResponseWriter, req *http.Request) {

	if s.rule.Debug {
		log.Printf("[DEBUG set_first] Rule: %+v, Request Host: %s, Headers: %v",
			s.rule, req.Host, req.Header)
	}

	for i, val := range s.rule.Values {

		headerValue := utils.GetValue(val, s.rule.HeaderPrefix, req)

		if s.rule.Debug {
			log.Printf("[DEBUG set_first] Trying value[%d] '%s': got headerValue='%s'",
				i, val, headerValue)
		}

		if headerValue != "" {

			if s.rule.Debug {
				log.Printf("[DEBUG set_first] MATCH! Setting %s='%s' (SetOnResponse=%v)",
					s.rule.Name, headerValue, s.rule.SetOnResponse)
			}

			if s.rule.SetOnResponse {
				rw.Header().Set(s.rule.Name, headerValue)

				return
			}

			if strings.EqualFold(s.rule.Header, "Host") {
				req.Host = headerValue
			} else {
				req.Header.Set(s.rule.Header, headerValue)
			}

			if s.rule.Debug {
				log.Printf("[DEBUG set_first] SUCCESS: Set %s='%s'", s.rule.Name, headerValue)
			}

			return
		}
	}

	if s.rule.Debug {
		log.Printf("[DEBUG set_first] No matching value found from %d attempts", len(s.rule.Values))
	}
}
