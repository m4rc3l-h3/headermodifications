package set

import (
	"net/http"

	"github.com/m4rc3l-h3/headermodifications/pkg/types"
	"github.com/m4rc3l-h3/headermodifications/pkg/utils"
	"github.com/m4rc3l-h3/headermodifications/pkg/utils/header"
)

type Set struct {
	rule *types.Rule
}

func (s *Set) Debug() bool {
	return s.rule.Debug
}

func New(rule types.Rule) (types.Handler, error) {
	return &Set{rule: &rule}, nil
}

func (s *Set) Validate() error {
	if s.rule.Header == "" || s.rule.Value == "" {
		return types.ErrMissingRequiredFields
	}

	return nil
}

func (s *Set) Handle(rw http.ResponseWriter, req *http.Request) (blocked bool) {

	newHeaderVal := utils.GetValue(s.rule.Value, s.rule.HeaderPrefix, req)

	if s.rule.SetOnResponse {
		rw.Header().Set(s.rule.Header, newHeaderVal)

		return false
	}

	header.Set(req, s.rule.Header, newHeaderVal)
	return false
}
