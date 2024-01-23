package linode

import (
	"reflect"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/mitchellh/mapstructure"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packersdk.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactId(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{"private/42", "packer-foobar", nil, generatedData}
	expected := "private/42"

	if a.Id() != expected {
		t.Fatalf("artifact ID should match: %v", expected)
	}
}

func TestArtifactString(t *testing.T) {
	generatedData := make(map[string]interface{})
	a := &Artifact{"private/42", "packer-foobar", nil, generatedData}
	expected := "Linode image: packer-foobar (private/42)"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	if result != expectedData {
		t.Fatalf("Bad: State data was %s instead of %s", result, expectedData)
	}

	// Invalid state
	result = artifact.State("invalid_key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for invalid state data name")
	}

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	if result != nil {
		t.Fatalf("Bad: State should be nil for nil StateData")
	}
}

func TestArtifactState_hcpPackerRegistryMetadata(t *testing.T) {
	region := "us-ord"
	artifact := &Artifact{
		ImageID:    "test-image",
		ImageLabel: "test-image-label",
		StateData: map[string]interface{}{
			"source_image": "linode/debian9",
			"region":       region,
			"linode_type":  "g6-nanode-1",
		},
	}
	// result should contain "something"
	result := artifact.State(registryimage.ArtifactStateURI)
	if result == nil {
		t.Fatalf("Bad: HCP Packer registry image data was nil")
	}

	// check for proper decoding of result into slice of registryimage.Image
	var image registryimage.Image
	err := mapstructure.Decode(result, &image)
	if err != nil {
		t.Errorf("Bad: unexpected error when trying to decode state into registryimage.Image %v", err)
	}

	// check that all properties of the images were set correctly
	expected := registryimage.Image{
		ImageID:        "test-image",
		ProviderName:   "linode",
		ProviderRegion: "us-ord",
		SourceImageID:  "linode/debian9",
		Labels: map[string]string{
			"source_image": "linode/debian9",
			"region":       region,
			"linode_type":  "g6-nanode-1",
		},
	}
	if !reflect.DeepEqual(image, expected) {
		t.Fatalf("Bad: expected %#v got %#v", expected, image)
	}
}
