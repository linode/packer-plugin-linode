package image

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/linode/packer-plugin-linode/helper/acceptance"
)

func TestImageDataSourceAcc_basic(t *testing.T) {
	if skip := acceptance.TestAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-image-data-source-basic",
		Type:     "linode",
		Template: testImageDataSourceAccBasic,
	})
}

const testImageDataSourceAccBasic = `
data "linode-image" "latest_ubuntu" {
    id_regex = "linode/ubuntu.*"
    latest = true
}

source "linode" "example" {
  image             = data.linode-image.latest_ubuntu.id
  instance_type     = "g6-nanode-1"
  region            = "us-mia"
  ssh_username      = "root"
}

build {
  sources = ["source.linode.example"]
}
`
