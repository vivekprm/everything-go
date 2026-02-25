// Package mid provides app level middleware support.
package mid

import (
	"ultimate-software-design/chat/foundation/web"
)

func checkIsError(e web.Encoder) error {
	err, hasError := e.(error)
	if hasError {
		return err
	}

	return nil
}

