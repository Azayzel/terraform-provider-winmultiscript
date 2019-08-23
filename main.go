package main

import (
	"terraform-provider-winmultiscript/winmultiscript"

	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{

		ProviderFunc: winmultiscript.Provider})
}
