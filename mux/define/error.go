package define

import (
	"encoding/json"
	"net/url"
)

type Errors url.Values

func (errs Errors) Error() string {
	return errs.String()
}

func (errs Errors) String() string {
	if len(errs) > 0 {
		v, _ := json.Marshal(errs)
		return string(v)
	}
	return ""
}

func (errs Errors) IsEmpty() bool {
	return len(errs) == 0
}

func (errs Errors) Add(name string, message string) {
	errs[name] = append(errs[name], message)
}

func (errs Errors) Set(name string, message ...string) {
	if len(message) > 0 {
		errs[name] = message
	}
}
