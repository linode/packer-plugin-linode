package linode

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/linode/packer-plugin-linode/helper/acceptance"
)

func TestBuilderAcc_basic(t *testing.T) {
	if skip := acceptance.TestAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-builder-basic",
		Type:     "linode",
		Template: testBuilderAccBasic,
	})
}

const testBuilderAccBasic = `
source "linode" "example" {
	image             = "linode/ubuntu22.04"
	instance_type     = "g6-nanode-1"
	region            = "us-mia"
	ssh_username      = "root"
}

build {
	sources = ["source.linode.example"]
}
`
