package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/stuttgart-things/terraform-provider-clusterbook/internal/provider"
)

func main() {
	if err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/stuttgart-things/clusterbook",
	}); err != nil {
		log.Fatal(err)
	}
}
