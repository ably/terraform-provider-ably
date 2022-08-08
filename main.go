package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"log"
	ably_control "terraform-provider-ably/internal/provider"
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
