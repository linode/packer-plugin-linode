package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/linode/packer-plugin-linode/helper"
)

func TestAccPreCheck(t *testing.T) bool {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar)
		return true
	}

	if v := os.Getenv(helper.TokenEnvVar); v == "" {
		t.Fatalf("%q must be set for acceptance tests", helper.TokenEnvVar)
		return true
	}
	return false
}
