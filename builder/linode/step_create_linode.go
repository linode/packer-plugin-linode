package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepCreateLinode struct {
	client linodego.Client
}

func (s *stepCreateLinode) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	handleError := func(prefix string, err error) multistep.StepAction {
		return errorHelper(state, ui, prefix, err)
	}

	ui.Say("Creating Linode...")

	createOpts := linodego.InstanceCreateOptions{
		RootPass:       c.Comm.Password(),
		AuthorizedKeys: []string{},
		Region:         c.Region,
		Type:           c.InstanceType,
		Label:          c.Label,
		Image:          c.Image,
		SwapSize:       &c.SwapSize,
	}

	if pubKey := string(c.Comm.SSHPublicKey); pubKey != "" {
		createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, pubKey)
	}

	createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, c.AuthorizedKeys...)

	instance, err := s.client.CreateInstance(ctx, createOpts)
	if err != nil {
		return handleError("Failed to create Linode Instance", err)
	}
	state.Put("instance", instance)
	state.Put("instance_id", instance.ID)

	// wait until instance is running
	instance, err = s.client.WaitForInstanceStatus(ctx, instance.ID, linodego.InstanceRunning, int(c.StateTimeout.Seconds()))
	if err != nil {
		return handleError("Failed to wait for Linode ready", err)
	}
	state.Put("instance", instance)

	disk, err := s.findDisk(ctx, instance.ID)
	if err != nil {
		return handleError("Failed to find instance disk", err)
	}

	if disk == nil {
		return handleError("Failed to find instance disk", errors.New("no suitable disk was found"))
	}
	state.Put("disk", disk)
	return multistep.ActionContinue
}

func (s *stepCreateLinode) findDisk(ctx context.Context, instanceID int) (*linodego.InstanceDisk, error) {
	disks, err := s.client.ListInstanceDisks(ctx, instanceID, nil)
	if err != nil {
		return nil, err
	}
	for _, disk := range disks {
		if disk.Filesystem != linodego.FilesystemSwap {
			return &disk, nil
		}
	}
	return nil, nil
}

func (s *stepCreateLinode) Cleanup(state multistep.StateBag) {
	instance, ok := state.GetOk("instance")
	if !ok {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	if err := s.client.DeleteInstance(context.Background(), instance.(*linodego.Instance).ID); err != nil {
		ui.Error("Error cleaning up Linode: " + err.Error())
	}
}
