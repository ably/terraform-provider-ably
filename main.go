package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ably_control "github.com/ably/terraform-provider-ably/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// VERSION is the version of this provider.
//
// It defaults to "dev" but is overridden during the release process using
// ldflags (e.g. go build -ldflags="-X main.VERSION=x.y.z").
var VERSION = "dev"

func main() {
	// print the version and exit if argv[1] is "version"
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("terraform-provider-ably " + VERSION)
		os.Exit(0)
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/ably/terraform-provider-ably",
	}

	err := providerserver.Serve(context.Background(), ably_control.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
