package main

import (
	"context"
	"flag"
	"log"

	"github.com/fi-ts/terraform-provider-fco/internal/provider"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"

func main() {
	flag.BoolVar(&shared.Debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/fi-ts/fco",
		Debug:   shared.Debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
