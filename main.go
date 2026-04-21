package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/dynu/terraform-provider-dynu/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New("dev"), providerserver.ServeOpts{
		Address: "registry.terraform.io/dynu/dynu",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
