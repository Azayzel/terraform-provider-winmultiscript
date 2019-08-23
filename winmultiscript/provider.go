package winmultiscript

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider Main Provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"winmultiscript": dataSourceFiles(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"winmultiscript": schema.DataSourceResourceShim(
				"winmultiscript",
				dataSourceFiles(),
			),
		},
	}
}
