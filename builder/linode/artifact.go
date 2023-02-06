// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package linode

import (
	"context"
	"fmt"
	"log"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/linode/linodego"
)

type Artifact struct {
	ImageID    string
	ImageLabel string

	Driver *linodego.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a Artifact) BuilderId() string { return BuilderID }
func (a Artifact) Files() []string   { return nil }
func (a Artifact) Id() string        { return a.ImageID }

func (a Artifact) String() string {
	return fmt.Sprintf("Linode image: %s (%s)", a.ImageLabel, a.ImageID)
}

func (a Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		return a.stateHCPPackerRegistryMetadata()
	}
	return a.StateData[name]
}

func (a Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.ImageID, a.ImageLabel)
	err := a.Driver.DeleteImage(context.TODO(), a.ImageID)
	return err
}

func (a Artifact) stateHCPPackerRegistryMetadata() interface{} {
	// create labels map
	labels := make(map[string]string)
	// get and set sourceImage from stateData into labels
	sourceImage, ok := a.StateData["source_image"].(string)
	if ok {
		labels["source_image"] = sourceImage
	}
	// get and set region from stateData into labels
	region, ok := a.StateData["region"].(string)
	if ok {
		labels["region"] = region
	}
	// get and set instance_type (specs) from stateData into labels
	linodeType, ok := a.StateData["linode_type"].(string)
	if ok {
		labels["linode_type"] = linodeType
	}
	// create the image from artifact
	image, err := registryimage.FromArtifact(a,
		registryimage.WithProvider("linode"),
		registryimage.WithID(a.ImageID),
		registryimage.WithSourceID(sourceImage),
		registryimage.WithRegion(region))
	if err != nil {
		log.Printf("[DEBUG] error encountered when creating registry image %s", err)
		return nil
	}
	image.Labels = labels
	return image
}
