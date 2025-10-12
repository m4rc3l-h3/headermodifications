package set

import (
	"net/http"

	"github.com/tomMoulard/htransformation/pkg/types"
	"github.com/tomMoulard/htransformation/pkg/utils"
	"github.com/tomMoulard/htransformation/pkg/utils/header"
)

type Set struct {
	rule *types.Rule
}

func New(rule types.Rule) (types.Handler, error) {
	return &Set{rule: &rule}, nil
}

func (s *Set) Validate() error {
	if s.rule.Header == "" ||
		(s.rule.HeaderPrefix != "" && s.rule.Value == "") ||
		(s.rule.Value != "" && s.rule.HeaderPrefix == "") {
		return types.ErrMissingRequiredFields
	}

	return nil
}

func (s *Set) Handle(rw http.ResponseWriter, req *http.Request) {

	newHeaderVal := utils.GetValue(s.rule.Value, s.rule.HeaderPrefix, req)

	if s.rule.SetOnResponse {
		rw.Header().Set(s.rule.Header, newHeaderVal)

		return
	}

	header.Set(req, s.rule.Header, newHeaderVal)
}
