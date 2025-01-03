package linode

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
	"github.com/linode/packer-plugin-linode/helper"
)

type stepCreateImage struct {
	client *linodego.Client
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	disk := state.Get("disk").(*linodego.InstanceDisk)
	instance := state.Get("instance").(*linodego.Instance)

	handleError := func(prefix string, err error) multistep.StepAction {
		return helper.ErrorHelper(state, ui, prefix, err)
	}

	creationPoller, err := s.client.NewEventPoller(
		ctx, instance.ID, linodego.EntityLinode, linodego.ActionDiskImagize)
	if err != nil {
		return handleError("Failed to create event poller", err)
	}

	ui.Say("Creating image...")
	image, err := s.client.CreateImage(ctx, linodego.ImageCreateOptions{
		DiskID:      disk.ID,
		Label:       c.ImageLabel,
		Description: c.Description,
		CloudInit:   c.CloudInit,
	})
	if err != nil {
		return handleError("Failed to create image", err)
	}

	_, err = creationPoller.WaitForFinished(ctx, int(c.ImageCreateTimeout.Seconds()))
	if err != nil {
		return handleError("Failed to wait for image creation", err)
	}

	if len(c.ImageRegions) > 0 {
		image, err = s.client.ReplicateImage(ctx, image.ID, linodego.ImageReplicateOptions{
			Regions: c.ImageRegions,
		})
		if err != nil {
			return handleError("Failed to replicate the image", err)
		}

		for _, r := range c.ImageRegions {
			_, err = s.client.WaitForImageRegionStatus(ctx, image.ID, r, linodego.ImageRegionStatusAvailable)
			if err != nil {
				return handleError("Failed to wait for the image replication", err)
			}
		}
	}

	image, err = s.client.GetImage(ctx, image.ID)
	if err != nil {
		return handleError("Failed to get image", err)
	}

	state.Put("image", image)
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}
