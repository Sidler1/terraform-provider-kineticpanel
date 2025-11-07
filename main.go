package main

import (
	"context"
	"flag"
	"log"

	"github.com/Sidler1/terraform-provider-kineticpanel/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "0.1.1-dev"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run with debugging")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address:         "registry.terraform.io/Sidler1/kineticpanel",
		ProtocolVersion: 6,
		Debug:           debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
