package linode

import (
	"context"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepCreateImage struct {
	client linodego.Client
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	disk := state.Get("disk").(*linodego.InstanceDisk)
	instance := state.Get("instance").(*linodego.Instance)

	handleError := func(prefix string, err error) multistep.StepAction {
		return errorHelper(state, ui, prefix, err)
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
	})

	if err != nil {
		return handleError("Failed to create image", err)
	}

	_, err = creationPoller.WaitForFinished(ctx, int(c.ImageCreateTimeout.Seconds()))
	if err != nil {
		return handleError("Failed to wait for image creation", err)
	}

	image, err = s.client.GetImage(ctx, image.ID)
	if err != nil {
		return handleError("Failed to get image", err)
	}

	state.Put("image", image)
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}
