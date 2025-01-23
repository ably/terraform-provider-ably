package provider

import (
	control "github.com/ably/ably-control-go"
)

func is404(err error) bool {
	e, ok := err.(control.ErrorInfo)
	return ok && e.StatusCode == 404
}
