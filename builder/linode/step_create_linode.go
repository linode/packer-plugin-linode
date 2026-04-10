package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
	"github.com/linode/packer-plugin-linode/helper"
)

type stepCreateLinode struct {
	client *linodego.Client
}

func flattenConfigInterfaceIPv4(i *InterfaceIPv4) *linodego.VPCIPv4 {
	if i == nil {
		return nil
	}

	return &linodego.VPCIPv4{
		VPC:     i.VPC,
		NAT1To1: i.NAT1To1,
	}
}

func flattenConfigInterface(i Interface) linodego.InstanceConfigInterfaceCreateOptions {
	return linodego.InstanceConfigInterfaceCreateOptions{
		IPAMAddress: i.IPAMAddress,
		Label:       i.Label,
		Purpose:     linodego.ConfigInterfacePurpose(i.Purpose),
		Primary:     i.Primary,
		SubnetID:    i.SubnetID,
		IPv4:        flattenConfigInterfaceIPv4(i.IPv4),
		IPRanges:    i.IPRanges,
	}
}

func flattenPublicInterface(public *PublicInterface) *linodego.PublicInterfaceCreateOptions {
	if public == nil {
		return nil
	}
	result := &linodego.PublicInterfaceCreateOptions{}
	if public.IPv4 != nil {
		addresses := make([]linodego.PublicInterfaceIPv4AddressCreateOptions, len(public.IPv4.Addresses))
		for i, addr := range public.IPv4.Addresses {
			addresses[i] = linodego.PublicInterfaceIPv4AddressCreateOptions{
				Address: addr.Address,
				Primary: addr.Primary,
			}
		}
		result.IPv4 = &linodego.PublicInterfaceIPv4CreateOptions{
			Addresses: linodego.Pointer(addresses),
		}
	}
	if public.IPv6 != nil {
		ranges := make([]linodego.PublicInterfaceIPv6RangeCreateOptions, len(public.IPv6.Ranges))
		for i, r := range public.IPv6.Ranges {
			ranges[i] = linodego.PublicInterfaceIPv6RangeCreateOptions{
				Range: r.Range,
			}
		}
		result.IPv6 = &linodego.PublicInterfaceIPv6CreateOptions{
			Ranges: linodego.Pointer(ranges),
		}
	}
	return result
}

func flattenVPCInterface(vpc *VPCInterface) *linodego.VPCInterfaceCreateOptions {
	if vpc == nil {
		return nil
	}
	result := &linodego.VPCInterfaceCreateOptions{
		SubnetID: vpc.SubnetID,
	}
	if vpc.IPv4 != nil {
		addresses := make([]linodego.VPCInterfaceIPv4AddressCreateOptions, len(vpc.IPv4.Addresses))
		ranges := make([]linodego.VPCInterfaceIPv4RangeCreateOptions, len(vpc.IPv4.Ranges))
		for i, addr := range vpc.IPv4.Addresses {
			addresses[i] = linodego.VPCInterfaceIPv4AddressCreateOptions{
				Address:        addr.Address,
				Primary:        addr.Primary,
				NAT1To1Address: addr.NAT1To1Address,
			}
		}
		for i, r := range vpc.IPv4.Ranges {
			ranges[i] = linodego.VPCInterfaceIPv4RangeCreateOptions{
				Range: r.Range,
			}
		}
		result.IPv4 = &linodego.VPCInterfaceIPv4CreateOptions{
			Addresses: linodego.Pointer(addresses),
			Ranges:    linodego.Pointer(ranges),
		}
	}
	return result
}

func flattenVLANInterface(vlan *VLANInterface) *linodego.VLANInterface {
	if vlan == nil {
		return nil
	}
	result := &linodego.VLANInterface{
		VLANLabel: vlan.VLANLabel,
	}
	if vlan.IPAMAddress != nil {
		result.IPAMAddress = vlan.IPAMAddress
	}
	return result
}

func flattenLinodeInterface(li LinodeInterface) (opts linodego.LinodeInterfaceCreateOptions) {
	opts.FirewallID = li.FirewallID

	if li.DefaultRoute != nil {
		opts.DefaultRoute = &linodego.InterfaceDefaultRoute{
			IPv4: li.DefaultRoute.IPv4,
			IPv6: li.DefaultRoute.IPv6,
		}
	}

	opts.Public = flattenPublicInterface(li.Public)
	opts.VPC = flattenVPCInterface(li.VPC)
	opts.VLAN = flattenVLANInterface(li.VLAN)

	return
}

func flattenMetadata(m Metadata) *linodego.InstanceMetadataOptions {
	if m.UserData == "" {
		return nil
	}

	return &linodego.InstanceMetadataOptions{
		UserData: m.UserData,
	}
}

func (s *stepCreateLinode) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	handleError := func(prefix string, err error) multistep.StepAction {
		return helper.ErrorHelper(state, ui, prefix, err)
	}

	ui.Say("Creating Linode...")

	// Determine if we're using custom disks/configs (explicit provisioning)
	// When custom disks are specified, we don't set the Image on instance creation
	// because we'll create disks and configs manually in a separate step.
	useCustomDisks := len(c.Disks) > 0

	createOpts := linodego.InstanceCreateOptions{
		PrivateIP:           c.PrivateIP,
		Region:              c.Region,
		Type:                c.InstanceType,
		Label:               c.Label,
		Tags:                c.Tags,
		FirewallID:          c.FirewallID,
		Metadata:            flattenMetadata(c.Metadata),
		InterfaceGeneration: linodego.InterfaceGeneration(c.InterfaceGeneration),
	}

	// Only set image-related options when NOT using custom disks
	if !useCustomDisks {
		createOpts.RootPass = c.Comm.Password()
		createOpts.Image = c.Image
		createOpts.SwapSize = c.SwapSize
		createOpts.StackScriptID = c.StackScriptID
		createOpts.StackScriptData = c.StackScriptData

		if pubKey := string(c.Comm.SSHPublicKey); pubKey != "" {
			createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, pubKey)
		}

	} else {
		ui.Say("Using custom disk configuration - instance will be created without an image")

		// When using custom disks, we need to boot the instance ourselves after config is created
		createOpts.Booted = linodego.Pointer(false)
	}

	interfaces := make([]linodego.InstanceConfigInterfaceCreateOptions, len(c.Interfaces))
	for i, v := range c.Interfaces {
		interfaces[i] = flattenConfigInterface(v)
	}

	linodeInterfaces := make([]linodego.LinodeInterfaceCreateOptions, len(c.LinodeInterfaces))
	for i, v := range c.LinodeInterfaces {
		linodeInterfaces[i] = flattenLinodeInterface(v)
	}

	// Only add legacy interfaces to instance creation when NOT using custom disks
	// (when using custom disks, legacy interfaces should be specified in the config block)
	// linode_interface (newer system) can be specified at instance level regardless of disk mode
	if !useCustomDisks && len(interfaces) > 0 {
		createOpts.Interfaces = interfaces
	}

	if len(linodeInterfaces) > 0 {
		createOpts.LinodeInterfaces = linodeInterfaces
	}

	createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, c.AuthorizedKeys...)
	createOpts.AuthorizedUsers = append(createOpts.AuthorizedUsers, c.AuthorizedUsers...)

	instance, err := s.client.CreateInstance(ctx, createOpts)
	if err != nil {
		return handleError("Failed to create Linode Instance", err)
	}
	state.Put("instance", instance)
	state.Put("instance_id", instance.ID)

	// When using custom disks, we skip waiting for running state here
	// because the instance won't boot until we create disks and configs
	if useCustomDisks {
		// Wait for instance to be in offline state (resources allocated)
		instance, err = s.client.WaitForInstanceStatus(ctx, instance.ID, linodego.InstanceOffline, int(c.StateTimeout.Seconds()))
		if err != nil {
			return handleError("Failed to wait for Linode to be offline", err)
		}
		state.Put("instance", instance)
		// Disk will be set by stepCreateDiskConfig
		return multistep.ActionContinue
	}

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
