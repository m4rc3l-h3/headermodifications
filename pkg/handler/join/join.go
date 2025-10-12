package join

import (
	"net/http"
	"strings"

	"github.com/m4rc3l-h3/htransformation/pkg/types"
	"github.com/m4rc3l-h3/htransformation/pkg/utils"
)

type Join struct {
	rule *types.Rule
}

func New(rule types.Rule) (types.Handler, error) {
	return &Join{rule: &rule}, nil
}

func (j *Join) Validate() error {
	if len(j.rule.Values) == 0 || j.rule.Sep == "" {
		return types.ErrMissingRequiredFields
	}

	return nil
}

func (j *Join) Handle(rw http.ResponseWriter, req *http.Request) {
	var val []string
	if strings.EqualFold(j.rule.Header, "Host") {
		val = []string{req.Host}
	} else {
		var ok bool
		val, ok = req.Header[j.rule.Header]

		if !ok {
			return
		}
	}

	newHeaderVal := val[0]
	for _, value := range j.rule.Values {
		newHeaderVal += j.rule.Sep + utils.GetValue(value, j.rule.HeaderPrefix, req)
	}

	if j.rule.SetOnResponse {
		rw.Header().Set(j.rule.Name, newHeaderVal)

		return
	}

	if strings.EqualFold(j.rule.Header, "Host") {
		req.Host = newHeaderVal
	} else {
		req.Header.Set(j.rule.Header, newHeaderVal)
	}
}
