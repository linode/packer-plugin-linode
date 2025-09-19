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
		result.IPv4 = &linodego.PublicInterfaceIPv4CreateOptions{
			Addresses: make([]linodego.PublicInterfaceIPv4AddressCreateOptions, len(public.IPv4.Addresses)),
		}
		for i, addr := range public.IPv4.Addresses {
			result.IPv4.Addresses[i] = linodego.PublicInterfaceIPv4AddressCreateOptions{
				Address: addr.Address,
				Primary: addr.Primary,
			}
		}
	}
	if public.IPv6 != nil {
		result.IPv6 = &linodego.PublicInterfaceIPv6CreateOptions{
			Ranges: make([]linodego.PublicInterfaceIPv6RangeCreateOptions, len(public.IPv6.Ranges)),
		}
		for i, r := range public.IPv6.Ranges {
			result.IPv6.Ranges[i] = linodego.PublicInterfaceIPv6RangeCreateOptions{
				Range: r.Range,
			}
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
		result.IPv4 = &linodego.VPCInterfaceIPv4CreateOptions{
			Addresses: make([]linodego.VPCInterfaceIPv4AddressCreateOptions, len(vpc.IPv4.Addresses)),
			Ranges:    make([]linodego.VPCInterfaceIPv4RangeCreateOptions, len(vpc.IPv4.Ranges)),
		}
		for i, addr := range vpc.IPv4.Addresses {
			result.IPv4.Addresses[i] = linodego.VPCInterfaceIPv4AddressCreateOptions{
				Address: addr.Address,
			}
		}
		for i, r := range vpc.IPv4.Ranges {
			result.IPv4.Ranges[i] = linodego.VPCInterfaceIPv4RangeCreateOptions{
				Range: r.Range,
			}
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
	opts.FirewallID = linodego.Pointer(li.FirewallID)

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

	createOpts := linodego.InstanceCreateOptions{
		RootPass:            c.Comm.Password(),
		AuthorizedKeys:      []string{},
		AuthorizedUsers:     []string{},
		PrivateIP:           c.PrivateIP,
		Region:              c.Region,
		StackScriptID:       c.StackScriptID,
		StackScriptData:     c.StackScriptData,
		Type:                c.InstanceType,
		Label:               c.Label,
		Image:               c.Image,
		SwapSize:            &c.SwapSize,
		Tags:                c.Tags,
		FirewallID:          c.FirewallID,
		Metadata:            flattenMetadata(c.Metadata),
		InterfaceGeneration: linodego.InterfaceGeneration(c.InterfaceGeneration),
	}

	interfaces := make([]linodego.InstanceConfigInterfaceCreateOptions, len(c.Interfaces))
	for i, v := range c.Interfaces {
		interfaces[i] = flattenConfigInterface(v)
	}

	linodeInterfaces := make([]linodego.LinodeInterfaceCreateOptions, len(c.LinodeInterfaces))
	for i, v := range c.LinodeInterfaces {
		linodeInterfaces[i] = flattenLinodeInterface(v)
	}

	if len(interfaces) > 0 {
		createOpts.Interfaces = interfaces
	}

	if len(linodeInterfaces) > 0 {
		createOpts.LinodeInterfaces = linodeInterfaces
	}

	if pubKey := string(c.Comm.SSHPublicKey); pubKey != "" {
		createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, pubKey)
	}

	createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, c.AuthorizedKeys...)
	createOpts.AuthorizedUsers = append(createOpts.AuthorizedUsers, c.AuthorizedUsers...)

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
