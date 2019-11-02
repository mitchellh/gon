package notarize

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// rawError is the error struct type for plist values.
type rawError struct {
	Code     int64             `plist:"code"`
	Message  string            `plist:"message"`
	UserInfo map[string]string `plist:"userInfo"`
}

// errorList turns a list of rawErrors into a Go error.
func errorList(errs []rawError) error {
	var result error
	for _, e := range errs {
		result = multierror.Append(result, fmt.Errorf("%s (%d)", e.Message, e.Code))
	}

	return result
}
