package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ably_control "github.com/ably/terraform-provider-ably/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// VERSION is the version of this provider.
//
// It defaults to 0.0.0 but is overridden during the release process using
// ldflags (e.g. go build -ldflags="-X main.VERSION=x.y.z").
//
// It is appended to the Ably-Agent HTTP header sent by the underlying Control
// API client (e.g. 'Ably-Agent: ably-control-go/1.0 terraform-provider-ably/x.y.z').
var VERSION = "0.0.0"

func main() {
	// print the version and exit if argv[1] is "version"
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("terraform-provider-ably " + VERSION)
		os.Exit(0)
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/ably/terraform-provider-ably",
	}

	err := providerserver.Serve(context.Background(), newProvider, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func newProvider() provider.Provider {
	return ably_control.New(VERSION)
}
