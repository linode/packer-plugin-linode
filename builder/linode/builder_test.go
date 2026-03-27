package linode

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

func testConfig() map[string]any {
	return map[string]any{
		"linode_token":  "bar",
		"region":        "us-ord",
		"instance_type": "g6-nanode-1",
		"ssh_username":  "root",
		"image":         "linode/arch",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw any
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]any{
		"linode_token": []string{},
	}

	_, warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Region(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	delete(config, "region")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should error")
	}

	expected := "us-ord"

	// Test set
	config["region"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Region != expected {
		t.Errorf("found %s, expected %s", b.config.Region, expected)
	}
}

func TestBuilderPrepare_Size(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	delete(config, "instance_type")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should error")
	}

	expected := "g6-nanode-1"

	// Test set
	config["instance_type"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.InstanceType != expected {
		t.Errorf("found %s, expected %s", b.config.InstanceType, expected)
	}
}

func TestBuilderPrepare_SwapSize(t *testing.T) {
	t.Run("omitted remains nil", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "swap_size")

		_, warnings, err := b.Prepare(config)
		if len(warnings) > 0 {
			t.Fatalf("bad: %#v", warnings)
		}
		if err != nil {
			t.Fatalf("should not have error: %s", err)
		}

		if b.config.SwapSize != nil {
			t.Fatalf("swap_size = %v, want nil", b.config.SwapSize)
		}
	})

	t.Run("explicit zero remains non-nil", func(t *testing.T) {
		var b Builder
		config := testConfig()
		config["swap_size"] = 0

		_, warnings, err := b.Prepare(config)
		if len(warnings) > 0 {
			t.Fatalf("bad: %#v", warnings)
		}
		if err != nil {
			t.Fatalf("should not have error: %s", err)
		}

		if b.config.SwapSize == nil || *b.config.SwapSize != 0 {
			t.Fatalf("swap_size = %v, want pointer to 0", b.config.SwapSize)
		}
	})
}

func TestBuilderPrepare_Image(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	delete(config, "image")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should error")
	}

	expected := "linode/debian12"

	// Test set
	config["image"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Image != expected {
		t.Errorf("found %s, expected %s", b.config.Image, expected)
	}
}

func TestBuilderPrepare_ImageLabel(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ImageLabel == "" {
		t.Errorf("invalid: %s", b.config.ImageLabel)
	}

	// Test set
	config["image_label"] = "foobarbaz"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test set with template
	config["image_label"] = "{{timestamp}}"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	_, err = strconv.ParseInt(b.config.ImageLabel, 0, 0)
	if err != nil {
		t.Fatalf("failed to parse int in template: %s", err)
	}
}

func TestBuilderPrepare_Label(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Label == "" {
		t.Errorf("invalid: %s", b.config.Label)
	}

	// Test normal set
	config["instance_label"] = "foobar"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test with template
	config["instance_label"] = "foobar-{{timestamp}}"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test with bad template
	config["instance_label"] = "foobar-{{"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_StateTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.StateTimeout != 5*time.Minute {
		t.Errorf("invalid: %s", b.config.StateTimeout)
	}

	// Test set
	config["state_timeout"] = "5m"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["state_timeout"] = "tubes"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ImageCreateTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ImageCreateTimeout != 10*time.Minute {
		t.Errorf("invalid: %s", b.config.ImageCreateTimeout)
	}

	// Test set
	config["image_create_timeout"] = "20m"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["image_create_timeout"] = "tubes"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_AuthorizedKeysAndUsers(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "authorized_keys")
	delete(config, "authorized_users")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKeys := []string{
		"ssh-rsa test@test",
	}

	expectedUsers := []string{
		"my_user1",
		"my_user2",
	}

	// Test set
	config["authorized_keys"] = expectedKeys
	config["authorized_users"] = expectedUsers
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.AuthorizedKeys, expectedKeys) {
		t.Errorf("got %v, expected %v", b.config.AuthorizedKeys, expectedKeys)
	}
	if !reflect.DeepEqual(b.config.AuthorizedUsers, expectedUsers) {
		t.Errorf("got %v, expected %v", b.config.AuthorizedKeys, expectedUsers)
	}
}

func TestBuilderPrepare_PrivateIP(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "private_ip")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPrivateIP := true
	config["private_ip"] = expectedPrivateIP

	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.PrivateIP, expectedPrivateIP) {
		t.Errorf("got %v, expected %v", b.config.PrivateIP, expectedPrivateIP)
	}
}

func TestBuilderPrepare_StackScripts(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "stackscript_id")
	delete(config, "stackscript_data")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStackScriptID := 123
	expectedStackScriptData := map[string]string{"test_data": "test_value"}

	// Test set
	config["stackscript_id"] = expectedStackScriptID
	config["stackscript_data"] = expectedStackScriptData
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.StackScriptID, expectedStackScriptID) {
		t.Errorf(
			"got %v, expected %v",
			b.config.StackScriptID,
			expectedStackScriptID,
		)
	}
	if !reflect.DeepEqual(b.config.StackScriptData, expectedStackScriptData) {
		t.Errorf(
			"got %v, expected %v",
			b.config.StackScriptData,
			expectedStackScriptData,
		)
	}
}

func TestBuilderPrepare_ConfigNetworkInterfaces(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "interface")
	delete(config, "authorized_users")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subnetID := 12345

	anyStr := "any"

	expectedInterfaces := []Interface{
		{
			Purpose: "public",
			Primary: true,
		},
		{
			Purpose: "vlan",
			VLANInterfaceAttributes: VLANInterfaceAttributes{
				Label:       "vlan-1",
				IPAMAddress: "10.0.0.1/24",
			},
		},
		{
			Purpose: "vpc",
			VPCInterfaceAttributes: VPCInterfaceAttributes{
				SubnetID: &subnetID,
				IPRanges: []string{"10.0.0.3/32"},
				IPv4: &InterfaceIPv4{
					VPC:     "10.0.0.2",
					NAT1To1: &anyStr,
				},
			},
		},
	}

	// Test set
	config["interface"] = expectedInterfaces
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.Interfaces, expectedInterfaces) {
		t.Errorf("got %v, expected %v", b.config.Interfaces, expectedInterfaces)
	}
}

func TestBuilderPrepare_LinodeNetworkInterfaces(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "linode_interface")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedLinodeInterfaces := []LinodeInterface{
		{
			FirewallID: linodego.Pointer(123),
			DefaultRoute: &InterfaceDefaultRoute{
				IPv4: linodego.Pointer(true),
				IPv6: linodego.Pointer(true),
			},
			Public: &PublicInterface{
				IPv4: &PublicInterfaceIPv4{
					Addresses: []PublicInterfaceIPv4Address{
						{
							Address: linodego.Pointer("auto"),
							Primary: linodego.Pointer(true),
						},
					},
				},
				IPv6: &PublicInterfaceIPv6{
					Ranges: []PublicInterfaceIPv6Range{
						{
							Range: "/64",
						},
					},
				},
			},
		},
		{
			FirewallID: linodego.Pointer(123),
			DefaultRoute: &InterfaceDefaultRoute{
				IPv4: linodego.Pointer(false),
				IPv6: linodego.Pointer(false),
			},
			VPC: &VPCInterface{
				SubnetID: 12345,
				IPv4: &VPCInterfaceIPv4{
					Addresses: []VPCInterfaceIPv4Address{
						{
							Address:        linodego.Pointer("auto"),
							Primary:        linodego.Pointer(false),
							NAT1To1Address: linodego.Pointer("auto"),
						},
					},
				},
			},
		},
		{
			DefaultRoute: &InterfaceDefaultRoute{
				IPv4: linodego.Pointer(false),
				IPv6: linodego.Pointer(false),
			},
			VLAN: &VLANInterface{
				VLANLabel:   "vlan-1",
				IPAMAddress: linodego.Pointer("10.0.0.1/24"),
			},
		},
	}

	// Test set
	config["linode_interface"] = expectedLinodeInterfaces
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.LinodeInterfaces, expectedLinodeInterfaces) {
		t.Errorf("got %v, expected %v", b.config.LinodeInterfaces, expectedLinodeInterfaces)
	}
}

func TestBuilderPrepare_CloudInit(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	delete(config, "cloud_init")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error; got %v", err)
	}

	if b.config.CloudInit {
		t.Fatalf("expected default to be false; got true")
	}

	// Const to silence warnings
	const expected = true

	// Test set
	config["cloud_init"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.CloudInit != expected {
		t.Errorf("found %s, expected %t", b.config.Region, expected)
	}
}

func TestBuilderPrepare_MetadataTagsFirewallID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test optional
	delete(config, "firewall_id")
	delete(config, "metadata")
	delete(config, "instance_tags")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFirewallID := 123
	config["firewall_id"] = expectedFirewallID

	expectedUserData := "foo"
	expectedMetadata := Metadata{
		UserData: expectedUserData,
	}
	config["metadata"] = map[string]string{
		"user_data": expectedUserData,
	}

	expectedTags := []string{
		"foo",
		"bar=baz",
		":!@#$%^&*()_+-=[]\\{}|;'\",./<>?`~",
	}
	config["instance_tags"] = expectedTags

	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.FirewallID, expectedFirewallID) {
		t.Errorf("got %v, expected %v", b.config.FirewallID, expectedFirewallID)
	}

	if !reflect.DeepEqual(b.config.Metadata, expectedMetadata) {
		t.Errorf("got %v, expected %v", b.config.Metadata, expectedMetadata)
	}

	if !reflect.DeepEqual(b.config.Tags, expectedTags) {
		t.Errorf("got %v, expected %v", b.config.Tags, expectedTags)
	}
}

func TestBuilderPrepare_ImageShareGroupIDs(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "image_share_group_ids")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.config.ImageShareGroupIDs) != 0 {
		t.Errorf("expected nil or empty, got %v", b.config.ImageShareGroupIDs)
	}

	expected := []int{101, 202, 303}
	config["image_share_group_ids"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if !reflect.DeepEqual(b.config.ImageShareGroupIDs, expected) {
		t.Errorf("got %v, expected %v", b.config.ImageShareGroupIDs, expected)
	}
}

func TestBuilderPrepare_CustomDisks(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with custom disk
	config["disk"] = []map[string]any{
		{
			"label":      "boot",
			"size":       25000,
			"image":      "linode/arch",
			"filesystem": "ext4",
		},
		{
			"label":      "swap",
			"size":       512,
			"filesystem": "swap",
		},
	}

	// Add config block (required when using custom disks)
	config["config"] = []map[string]any{
		{
			"label":       "my-config",
			"kernel":      "linode/latest-64bit",
			"root_device": "/dev/sda",
			"devices": map[string]any{
				"sda": map[string]any{
					"disk_label": "boot",
				},
				"sdb": map[string]any{
					"disk_label": "swap",
				},
			},
		},
	}

	// When using custom disks, image should not be required at top level
	delete(config, "image")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error with custom disks: %s", err)
	}

	if len(b.config.Disks) != 2 {
		t.Errorf("expected 2 disks, got %d", len(b.config.Disks))
	}

	if b.config.Disks[0].Label != "boot" {
		t.Errorf("expected first disk label to be 'boot', got %s", b.config.Disks[0].Label)
	}

	if b.config.Disks[1].Filesystem != "swap" {
		t.Errorf("expected second disk filesystem to be 'swap', got %s", b.config.Disks[1].Filesystem)
	}

	if len(b.config.InstanceConfigs) != 1 {
		t.Errorf("expected 1 config, got %d", len(b.config.InstanceConfigs))
	}
}

func TestBuilderPrepare_CustomConfig(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test with custom config
	config["config"] = []map[string]any{
		{
			"label":       "my-config",
			"comments":    "boot config",
			"kernel":      "linode/latest-64bit",
			"root_device": "/dev/sda",
			"devices": map[string]any{
				"sda": map[string]any{
					"disk_label": "boot",
				},
			},
			"helpers": map[string]any{
				"updatedb_disabled":  true,
				"distro":             true,
				"modules_dep":        true,
				"network":            true,
				"devtmpfs_automount": true,
			},
		},
	}

	config["disk"] = []map[string]any{
		{
			"label": "boot",
			"size":  25000,
			"image": "linode/arch",
		},
	}

	// When using custom disks, image should not be required at top level
	delete(config, "image")

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error with custom config: %s", err)
	}

	if len(b.config.InstanceConfigs) != 1 {
		t.Errorf("expected 1 config, got %d", len(b.config.InstanceConfigs))
	}

	if b.config.InstanceConfigs[0].Label != "my-config" {
		t.Errorf("expected config label to be 'my-config', got %s", b.config.InstanceConfigs[0].Label)
	}

	if b.config.InstanceConfigs[0].Kernel != "linode/latest-64bit" {
		t.Errorf("expected kernel to be 'linode/latest-64bit', got %s", b.config.InstanceConfigs[0].Kernel)
	}

	if b.config.InstanceConfigs[0].Devices == nil {
		t.Error("expected devices to be set")
	}

	if b.config.InstanceConfigs[0].Devices.SDA == nil {
		t.Error("expected SDA device to be set")
	}

	if b.config.InstanceConfigs[0].Devices.SDA.DiskLabel != "boot" {
		t.Errorf("expected SDA disk_label to be 'boot', got %s", b.config.InstanceConfigs[0].Devices.SDA.DiskLabel)
	}
}

func TestBuilderPrepare_CustomDisksValidation(t *testing.T) {
	// Test that image and custom disks cannot be specified together
	t.Run("ImageAndDisks", func(t *testing.T) {
		var b Builder
		config := testConfig()
		config["image"] = "linode/arch"
		config["disk"] = []map[string]any{
			{
				"label":      "boot",
				"size":       25000,
				"image":      "linode/arch",
				"filesystem": "ext4",
			},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when specifying both image and custom disks")
		}
		if !strings.Contains(err.Error(), "cannot specify both image and custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	// Test that config is required when using custom disks
	t.Run("DisksWithoutConfig", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["disk"] = []map[string]any{
			{
				"label":      "boot",
				"size":       25000,
				"image":      "linode/arch",
				"filesystem": "ext4",
			},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when using custom disks without config")
		}
		if !strings.Contains(err.Error(), "disk and config blocks must be specified together") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	// Test that either image or disks must be specified
	t.Run("NoImageOrDisks", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when neither image nor disks specified")
		}
		if !strings.Contains(err.Error(), "either image or custom disks must be specified") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	// Test incompatible attributes with custom disks
	t.Run("IncompatibleAuthorizedKeys", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["authorized_keys"] = []string{"ssh-rsa test"}
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with authorized_keys and custom disks")
		}
		if !strings.Contains(err.Error(), "authorized_keys cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("IncompatibleAuthorizedUsers", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["authorized_users"] = []string{"user1"}
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with authorized_users and custom disks")
		}
		if !strings.Contains(err.Error(), "authorized_users cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("IncompatibleSwapSize", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["swap_size"] = 512
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with swap_size and custom disks")
		}
		if !strings.Contains(err.Error(), "swap_size cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("IncompatibleStackScriptID", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["stackscript_id"] = 12345
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with stackscript_id and custom disks")
		}
		if !strings.Contains(err.Error(), "stackscript_id cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("IncompatibleStackScriptData", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["stackscript_data"] = map[string]string{"key": "value"}
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with stackscript_data and custom disks")
		}
		if !strings.Contains(err.Error(), "stackscript_data cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("IncompatibleInterface", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["interface"] = []map[string]any{
			{"purpose": "public"},
		}
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error with interface and custom disks")
		}
		if !strings.Contains(err.Error(), "interface blocks cannot be specified when using custom disks") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("LinodeInterfaceWithCustomDisks", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		// linode_interface should be ALLOWED with custom disks
		config["linode_interface"] = []map[string]any{
			{"public": map[string]any{}},
		}
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config", "root_device": "/dev/sda", "devices": map[string]any{"sda": map[string]any{"disk_label": "boot"}}},
		}

		_, _, err := b.Prepare(config)
		if err != nil {
			t.Fatalf("linode_interface should be allowed with custom disks, got error: %s", err)
		}
	})

	t.Run("MissingRootDevice", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config", "devices": map[string]any{"sda": map[string]any{"disk_label": "boot"}}},
		}
		// Missing root_device in boot config - should default to /dev/sda

		_, _, err := b.Prepare(config)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		// Verify default was applied
		if b.config.InstanceConfigs[0].RootDevice != "/dev/sda" {
			t.Fatalf("expected root_device to default to /dev/sda, got: %s", b.config.InstanceConfigs[0].RootDevice)
		}
	})

	t.Run("RootDevicePointsToUndefinedDisk", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config", "root_device": "/dev/sda", "devices": map[string]any{"sda": map[string]any{"disk_label": "nonexistent"}}},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when root_device points to undefined disk")
		}
		if !strings.Contains(err.Error(), "root_device points to disk") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("ConfigWithoutDisks", func(t *testing.T) {
		var b Builder
		config := testConfig()
		// Specify config blocks without disk blocks
		config["config"] = []map[string]any{
			{"label": "my-config"},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when using config blocks without disk blocks")
		}
		if !strings.Contains(err.Error(), "disk and config blocks must be specified together") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("DuplicateDiskLabels", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["disk"] = []map[string]any{
			{"label": "boot", "size": 25000, "image": "linode/arch"},
			{"label": "boot", "size": 512, "filesystem": "swap"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config", "root_device": "/dev/sda", "devices": map[string]any{"sda": map[string]any{"disk_label": "boot"}}},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when disk labels are duplicated")
		}
		if !strings.Contains(err.Error(), "duplicate disk label \"boot\" found") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})

	t.Run("EmptyDiskLabel", func(t *testing.T) {
		var b Builder
		config := testConfig()
		delete(config, "image")
		config["disk"] = []map[string]any{
			{"label": "", "size": 25000, "image": "linode/arch"},
		}
		config["config"] = []map[string]any{
			{"label": "my-config", "root_device": "/dev/sda", "devices": map[string]any{"sda": map[string]any{"disk_label": ""}}},
		}

		_, _, err := b.Prepare(config)
		if err == nil {
			t.Fatal("expected error when disk label is empty")
		}
		if !strings.Contains(err.Error(), "disk label cannot be empty") {
			t.Fatalf("expected specific error message, got: %s", err)
		}
	})
}
