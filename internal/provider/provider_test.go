package ably_control

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func Provider() *schema.Provider {
	// Ably Provider requires an API Token.
	// The URL is optional and defaults to the prod control API endpoint
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"terraform-provider-ably": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ABLY_ACCOUNT_TOKEN"); v == "" {
		t.Fatal("ABLY_ACCOUNT_TOKEN must be set for acceptance tests")
	}
}
