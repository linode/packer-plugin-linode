package linode

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestBuilderAcc_basic(t *testing.T) {
	if skip := testAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-builder-basic",
		Type:     "linode",
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar))
		return true
	}

	if v := os.Getenv("LINODE_TOKEN"); v == "" {
		t.Fatal("LINODE_TOKEN must be set for acceptance tests")
		return true
	}
	return false
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "linode",
		"region": "us-east",
		"instance_type": "g6-nanode-1",
		"image": "linode/alpine3.9",
		"ssh_username": "root"
	}]
}
`
