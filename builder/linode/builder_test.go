package linode

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"linode_token":  "bar",
		"region":        "us-east",
		"instance_type": "g6-nanode-1",
		"ssh_username":  "root",
		"image":         "linode/alpine3.9",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
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

	expected := "us-east"

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

	expected := "linode/alpine3.9"

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

func TestBuilderPrepare_NetworkInterfaces(t *testing.T) {
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

	expectedInterfaces := []Interface{
		{
			Purpose: "public",
			Primary: true,
		},
		{
			Purpose:     "vlan",
			Label:       "vlan-1",
			IPAMAddress: "10.0.0.1/24",
		},
		{
			Purpose:  "vpc",
			SubnetID: &subnetID,
			IPv4: &InterfaceIPv4{
				VPC:     "10.0.0.2",
				NAT1To1: "any",
			},
			IPRanges: []string{"10.0.0.5/32"},
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
