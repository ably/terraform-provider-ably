package main

import (
	"context"
	"log"

	ably_control "github.com/ably/terraform-provider-ably/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/ably/terraform-provider-ably",
	}

	err := providerserver.Serve(context.Background(), ably_control.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
