package set_first

import (
	"net/http"
	"strings"

	"github.com/tomMoulard/htransformation/pkg/types"
	"github.com/tomMoulard/htransformation/pkg/utils"
)

type SetFirst struct {
	rule *types.Rule
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

	for _, val := range s.rule.Values {
		headerValue := utils.GetValue(val, s.rule.HeaderPrefix, req)
		if headerValue != "" {
			if s.rule.SetOnResponse {
				rw.Header().Set(s.rule.Name, headerValue)

				return
			}

			if strings.EqualFold(s.rule.Header, "Host") {
				req.Host = headerValue
			} else {
				req.Header.Set(s.rule.Header, headerValue)
			}

			return
		}
	}
}
