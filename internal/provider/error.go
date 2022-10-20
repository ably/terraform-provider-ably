package ably_control

import (
	ably_control_go "github.com/ably/ably-control-go"
)

func is_404(err error) bool {
	e, ok := err.(ably_control_go.ErrorInfo)
	return ok && e.StatusCode == 404
}
