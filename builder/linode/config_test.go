package linode

import (
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestPrepare(t *testing.T) {
	data := communicator.SSH{
		SSHUsername: "root",
		// more variables can be added depending on test scentarios
	}

	config := &Config{
		ctx:                 interpolate.Context{},
		Comm:                communicator.Config{SSH: data},
		PersonalAccessToken: "test-linode-access-token",
		Region:              "us-ord",
		InstanceType:        "g6-standard-1",
		Image:               "linode/debian10",
	}

	warnings, err := config.Prepare()
	if err != nil {
		t.Errorf("Prepare failed with error: %v", err)
	}

	if len(warnings) > 0 {
		t.Logf("Warnings during Prepare: %v", warnings)
	}

	expectedStateTimeout := 5 * time.Minute
	if config.StateTimeout != expectedStateTimeout {
		t.Errorf("Expected StateTimeout: %v, Got: %v", expectedStateTimeout, config.StateTimeout)
	}

	expectedImageCreateTimeout := 10 * time.Minute
	if config.ImageCreateTimeout != expectedImageCreateTimeout {
		t.Errorf("Expected ImageCreateTimeout: %v, Got: %v", expectedImageCreateTimeout, config.ImageCreateTimeout)
	}
}

func TestHCL2Spec(t *testing.T) {
	packerBuildName := "testBuildName"
	sshHost := "test-host"

	flatConfig := &FlatConfig{
		PackerBuildName: &packerBuildName,
		SSHHost:         &sshHost,
	}

	hclSpec := flatConfig.HCL2Spec()

	expectedAttributes := []string{
		"packer_build_name",
		"ssh_host",
	}

	for _, attr := range expectedAttributes {
		if _, exists := hclSpec[attr]; !exists {
			t.Errorf("Expected attribute %s in HCL spec, but it's missing", attr)
		}
	}
}

func TestHCL2SpecInterface(t *testing.T) {
	purpose := "eth0"
	label := "PrimaryInterfaceLabel"
	ipamAddress := "192.168.1.10"

	flatIface := &FlatInterface{
		Purpose:     &purpose,
		Label:       &label,
		IPAMAddress: &ipamAddress,
	}

	hclSpec := flatIface.HCL2Spec()

	expectedAttributes := []string{
		"purpose",
		"label",
		"ipam_address",
	}

	for _, attr := range expectedAttributes {
		if _, exists := hclSpec[attr]; !exists {
			t.Errorf("Expected attribute %s in HCL spec, but it's missing", attr)
		}
	}
}
