package ably_control

import (
	control "github.com/ably/ably-control-go"
)

func is_404(err error) bool {
	e, ok := err.(control.ErrorInfo)
	return ok && e.StatusCode == 404
}
