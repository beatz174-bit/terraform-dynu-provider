package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/dynu/terraform-provider-dynu/internal/provider"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("terraform-provider-dynu version=%s commit=%s built=%s\n", version, commit, date)
		return
	}

	log.Printf("starting terraform-provider-dynu version=%s commit=%s built=%s", version, commit, date)

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/dynu/dynu",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
