package set_from

import (
	"net/http"
	"strings"

	"github.com/tomMoulard/htransformation/pkg/types"
	"github.com/tomMoulard/htransformation/pkg/utils"
)

type SetFrom struct {
	rule *types.Rule
}

func New(rule types.Rule) (types.Handler, error) {
	return &SetFrom{rule: &rule}, nil
}

func (s *SetFrom) Validate() error {
	if s.rule.Value == "" || s.rule.HeaderPrefix == "" || s.rule.Header == "" {
		return types.ErrMissingRequiredFields
	}

	return nil
}

func (s *SetFrom) Handle(rw http.ResponseWriter, req *http.Request) {

	newHeaderVal := utils.GetValue(s.rule.Value, s.rule.HeaderPrefix, req)

	if s.rule.SetOnResponse {
		rw.Header().Set(s.rule.Name, newHeaderVal)

		return
	}

	if strings.EqualFold(s.rule.Header, "Host") {
		req.Host = newHeaderVal
	} else {
		req.Header.Set(s.rule.Header, newHeaderVal)
	}
}
