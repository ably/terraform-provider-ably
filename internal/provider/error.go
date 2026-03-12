// Package provider implements the Ably provider for Terraform
package provider

import (
	"errors"

	control "github.com/ably/terraform-provider-ably/client"
)

func is404(err error) bool {
	var e *control.Error
	return errors.As(err, &e) && e.StatusCode == 404
}
