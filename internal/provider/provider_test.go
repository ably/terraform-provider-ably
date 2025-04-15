package ably_control

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const providerConfig = `provider "ably" {}`

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ably": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ABLY_ACCOUNT_TOKEN"); v == "" {
		t.Fatal("ABLY_ACCOUNT_TOKEN must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	// Just test that the provider type exists
	p := &AblyProvider{}
	if p == nil {
		t.Fatal("Provider is nil")
	}
	// Validate the provider satisfies the interface
	var _ provider.Provider = &AblyProvider{}
}
