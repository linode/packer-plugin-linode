package linode

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
	"github.com/linode/packer-plugin-linode/helper"
)

// stepCreateDiskConfig creates custom disks and configuration profiles for a Linode instance.
// This step runs after the instance is created without an image (when custom disks/configs are specified).
type stepCreateDiskConfig struct {
	client *linodego.Client
}

func flattenDisk(d Disk) linodego.InstanceDiskCreateOptions {
	return linodego.InstanceDiskCreateOptions{
		Label:           d.Label,
		Size:            d.Size,
		Image:           d.Image,
		Filesystem:      d.Filesystem,
		AuthorizedKeys:  d.AuthorizedKeys,
		AuthorizedUsers: d.AuthorizedUsers,
		StackscriptID:   d.StackscriptID,
		StackscriptData: d.StackscriptData,
	}
}

func flattenInstanceConfigHelpers(h *InstanceConfigHelpers) *linodego.InstanceConfigHelpers {
	if h == nil {
		return nil
	}

	result := &linodego.InstanceConfigHelpers{}

	if h.UpdateDBDisabled != nil {
		result.UpdateDBDisabled = *h.UpdateDBDisabled
	}
	if h.Distro != nil {
		result.Distro = *h.Distro
	}
	if h.ModulesDep != nil {
		result.ModulesDep = *h.ModulesDep
	}
	if h.Network != nil {
		result.Network = *h.Network
	}
	if h.DevTmpFsAutomount != nil {
		result.DevTmpFsAutomount = *h.DevTmpFsAutomount
	}

	return result
}

// resolveDiskLabel resolves a disk label to a disk ID using the provided map.
func resolveDiskLabel(label string, diskLabelToID map[string]int) (int, error) {
	if label == "" {
		return 0, fmt.Errorf("disk label cannot be empty")
	}

	diskID, ok := diskLabelToID[label]
	if !ok {
		return 0, fmt.Errorf("disk with label %q not found", label)
	}
	return diskID, nil
}

// flattenInstanceConfigDevice resolves disk labels and creates the linodego device struct.
func flattenInstanceConfigDevice(d *InstanceConfigDevice, diskLabelToID map[string]int) (*linodego.InstanceConfigDevice, error) {
	if d == nil {
		return nil, nil
	}

	result := &linodego.InstanceConfigDevice{}

	if d.DiskLabel != "" {
		diskID, err := resolveDiskLabel(d.DiskLabel, diskLabelToID)
		if err != nil {
			return nil, err
		}
		result.DiskID = diskID
	}

	if d.VolumeID != 0 {
		result.VolumeID = d.VolumeID
	}

	// A device must have exactly one of disk or volume (not both, not neither)
	if (result.DiskID == 0) == (result.VolumeID == 0) {
		return nil, fmt.Errorf("device must specify exactly one of disk_label or volume_id")
	}

	return result, nil
}

// flattenInstanceConfigDevices resolves all device slots.
func flattenInstanceConfigDevices(d *InstanceConfigDevices, diskLabelToID map[string]int) (linodego.InstanceConfigDeviceMap, error) {
	result := linodego.InstanceConfigDeviceMap{}

	if d == nil {
		return result, nil
	}

	var err error

	// Define explicit mappings for all device slots (sda through sdbl)
	deviceMappings := []struct {
		name string
		src  *InstanceConfigDevice
		dst  **linodego.InstanceConfigDevice
	}{
		// sda through sdz
		{name: "sda", src: d.SDA, dst: &result.SDA},
		{name: "sdb", src: d.SDB, dst: &result.SDB},
		{name: "sdc", src: d.SDC, dst: &result.SDC},
		{name: "sdd", src: d.SDD, dst: &result.SDD},
		{name: "sde", src: d.SDE, dst: &result.SDE},
		{name: "sdf", src: d.SDF, dst: &result.SDF},
		{name: "sdg", src: d.SDG, dst: &result.SDG},
		{name: "sdh", src: d.SDH, dst: &result.SDH},
		{name: "sdi", src: d.SDI, dst: &result.SDI},
		{name: "sdj", src: d.SDJ, dst: &result.SDJ},
		{name: "sdk", src: d.SDK, dst: &result.SDK},
		{name: "sdl", src: d.SDL, dst: &result.SDL},
		{name: "sdm", src: d.SDM, dst: &result.SDM},
		{name: "sdn", src: d.SDN, dst: &result.SDN},
		{name: "sdo", src: d.SDO, dst: &result.SDO},
		{name: "sdp", src: d.SDP, dst: &result.SDP},
		{name: "sdq", src: d.SDQ, dst: &result.SDQ},
		{name: "sdr", src: d.SDR, dst: &result.SDR},
		{name: "sds", src: d.SDS, dst: &result.SDS},
		{name: "sdt", src: d.SDT, dst: &result.SDT},
		{name: "sdu", src: d.SDU, dst: &result.SDU},
		{name: "sdv", src: d.SDV, dst: &result.SDV},
		{name: "sdw", src: d.SDW, dst: &result.SDW},
		{name: "sdx", src: d.SDX, dst: &result.SDX},
		{name: "sdy", src: d.SDY, dst: &result.SDY},
		{name: "sdz", src: d.SDZ, dst: &result.SDZ},
		// sdaa through sdaz
		{name: "sdaa", src: d.SDAA, dst: &result.SDAA},
		{name: "sdab", src: d.SDAB, dst: &result.SDAB},
		{name: "sdac", src: d.SDAC, dst: &result.SDAC},
		{name: "sdad", src: d.SDAD, dst: &result.SDAD},
		{name: "sdae", src: d.SDAE, dst: &result.SDAE},
		{name: "sdaf", src: d.SDAF, dst: &result.SDAF},
		{name: "sdag", src: d.SDAG, dst: &result.SDAG},
		{name: "sdah", src: d.SDAH, dst: &result.SDAH},
		{name: "sdai", src: d.SDAI, dst: &result.SDAI},
		{name: "sdaj", src: d.SDAJ, dst: &result.SDAJ},
		{name: "sdak", src: d.SDAK, dst: &result.SDAK},
		{name: "sdal", src: d.SDAL, dst: &result.SDAL},
		{name: "sdam", src: d.SDAM, dst: &result.SDAM},
		{name: "sdan", src: d.SDAN, dst: &result.SDAN},
		{name: "sdao", src: d.SDAO, dst: &result.SDAO},
		{name: "sdap", src: d.SDAP, dst: &result.SDAP},
		{name: "sdaq", src: d.SDAQ, dst: &result.SDAQ},
		{name: "sdar", src: d.SDAR, dst: &result.SDAR},
		{name: "sdas", src: d.SDAS, dst: &result.SDAS},
		{name: "sdat", src: d.SDAT, dst: &result.SDAT},
		{name: "sdau", src: d.SDAU, dst: &result.SDAU},
		{name: "sdav", src: d.SDAV, dst: &result.SDAV},
		{name: "sdaw", src: d.SDAW, dst: &result.SDAW},
		{name: "sdax", src: d.SDAX, dst: &result.SDAX},
		{name: "sday", src: d.SDAY, dst: &result.SDAY},
		{name: "sdaz", src: d.SDAZ, dst: &result.SDAZ},
		// sdba through sdbl
		{name: "sdba", src: d.SDBA, dst: &result.SDBA},
		{name: "sdbb", src: d.SDBB, dst: &result.SDBB},
		{name: "sdbc", src: d.SDBC, dst: &result.SDBC},
		{name: "sdbd", src: d.SDBD, dst: &result.SDBD},
		{name: "sdbe", src: d.SDBE, dst: &result.SDBE},
		{name: "sdbf", src: d.SDBF, dst: &result.SDBF},
		{name: "sdbg", src: d.SDBG, dst: &result.SDBG},
		{name: "sdbh", src: d.SDBH, dst: &result.SDBH},
		{name: "sdbi", src: d.SDBI, dst: &result.SDBI},
		{name: "sdbj", src: d.SDBJ, dst: &result.SDBJ},
		{name: "sdbk", src: d.SDBK, dst: &result.SDBK},
		{name: "sdbl", src: d.SDBL, dst: &result.SDBL},
	}

	for _, mapping := range deviceMappings {
		if *mapping.dst, err = flattenInstanceConfigDevice(mapping.src, diskLabelToID); err != nil {
			return result, fmt.Errorf("%s: %w", mapping.name, err)
		}
	}

	return result, nil
}

// flattenInstanceConfig creates the linodego config create options.
func flattenInstanceConfig(cfg InstanceConfig, diskLabelToID map[string]int) (linodego.InstanceConfigCreateOptions, error) {
	devices, err := flattenInstanceConfigDevices(cfg.Devices, diskLabelToID)
	if err != nil {
		return linodego.InstanceConfigCreateOptions{}, fmt.Errorf("failed to resolve devices: %w", err)
	}

	// Flatten legacy interfaces if specified in the config block
	interfaces := make([]linodego.InstanceConfigInterfaceCreateOptions, len(cfg.Interfaces))
	for i, v := range cfg.Interfaces {
		interfaces[i] = flattenConfigInterface(v)
	}

	opts := linodego.InstanceConfigCreateOptions{
		Label:       cfg.Label,
		Comments:    cfg.Comments,
		Devices:     devices,
		Helpers:     flattenInstanceConfigHelpers(cfg.Helpers),
		Interfaces:  interfaces,
		MemoryLimit: cfg.MemoryLimit,
		Kernel:      cfg.Kernel,
		InitRD:      cfg.InitRD,
		RunLevel:    cfg.RunLevel,
		VirtMode:    cfg.VirtMode,
	}

	if cfg.RootDevice != "" {
		opts.RootDevice = &cfg.RootDevice
	}

	return opts, nil
}

// selectBootConfig determines which configuration profile should be booted.
// Returns the index of the config to boot, or an error if multiple configs have booted=true.
// If no configs have booted=true, returns 0 (first config).
// If one config has booted=true, returns its index.
func selectBootConfig(configs []InstanceConfig) (int, error) {
	bootedCount := 0
	bootedIndex := -1

	for i, cfg := range configs {
		if cfg.Booted {
			bootedCount++
			bootedIndex = i
		}
	}

	if bootedCount > 1 {
		return 0, errors.New("only one configuration profile can have 'booted' set to true")
	}

	if bootedIndex >= 0 {
		return bootedIndex, nil
	}

	return 0, nil
}

func (s *stepCreateDiskConfig) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	instance := state.Get("instance").(*linodego.Instance)

	handleError := func(prefix string, err error) multistep.StepAction {
		return helper.ErrorHelper(state, ui, prefix, err)
	}

	// Skip if no custom disks or configs are defined
	if len(c.Disks) == 0 && len(c.InstanceConfigs) == 0 {
		return multistep.ActionContinue
	}

	// Map to track disk label -> disk ID for config device resolution
	diskLabelToID := make(map[string]int)

	// Get the boot disk label from the boot config's root_device
	// This is validated during Prepare() so it should always succeed
	bootDiskLabel, err := c.getBootDiskLabel()
	if err != nil {
		return handleError("Failed to determine boot disk", err)
	}

	// Create disks
	for _, diskCfg := range c.Disks {
		ui.Say(fmt.Sprintf("Creating disk: %s...", diskCfg.Label))

		diskOpts := flattenDisk(diskCfg)

		// Only append SSH key to the disk specified by root_device (the boot disk)
		// Note: Top-level authorized_keys/authorized_users are validated to be empty when using custom disks
		if diskCfg.Label == bootDiskLabel {
			if len(c.Comm.SSHPublicKey) > 0 {
				diskOpts.AuthorizedKeys = append(diskOpts.AuthorizedKeys, string(c.Comm.SSHPublicKey))
			}
		}

		if diskOpts.RootPass == "" && diskOpts.Image != "" {
			diskOpts.RootPass = c.Comm.Password()
		}

		disk, err := s.client.CreateInstanceDisk(ctx, instance.ID, diskOpts)
		if err != nil {
			return handleError(fmt.Sprintf("Failed to create disk %q", diskCfg.Label), err)
		}

		// Wait for disk to be ready
		disk, err = s.client.WaitForInstanceDiskStatus(ctx, instance.ID, disk.ID, linodego.DiskReady, int(c.StateTimeout.Seconds()))
		if err != nil {
			return handleError(fmt.Sprintf("Failed to wait for disk %q", diskCfg.Label), err)
		}

		// Resolve a disk inconsistency where the disk may not be immediately bootable after creation
		time.Sleep(1 * time.Second)

		ui.Say(fmt.Sprintf("Disk %s created with ID: %d", disk.Label, disk.ID))
		diskLabelToID[disk.Label] = disk.ID
	}

	// Store disk map in state for other steps
	state.Put("disk_label_to_id", diskLabelToID)

	// Determine which config to boot
	bootConfigIndex, err := selectBootConfig(c.InstanceConfigs)
	if err != nil {
		return handleError("Multiple configuration profiles marked as booted", err)
	}

	// Create configuration profiles and track the boot config ID
	var bootConfigID int
	for i, cfgProfile := range c.InstanceConfigs {
		ui.Say(fmt.Sprintf("Creating configuration profile: %s...", cfgProfile.Label))

		configOpts, err := flattenInstanceConfig(cfgProfile, diskLabelToID)
		if err != nil {
			return handleError(fmt.Sprintf("Failed to prepare config %q", cfgProfile.Label), err)
		}

		config, err := s.client.CreateInstanceConfig(ctx, instance.ID, configOpts)
		if err != nil {
			return handleError(fmt.Sprintf("Failed to create config %q", cfgProfile.Label), err)
		}

		ui.Say(fmt.Sprintf("Configuration profile %s created with ID: %d", config.Label, config.ID))

		// Track the config that should be used for booting
		if i == bootConfigIndex {
			bootConfigID = config.ID
			if cfgProfile.Booted {
				ui.Say(fmt.Sprintf("Configuration profile %s will be used for booting", config.Label))
			}
		}
	}

	// Find the disk for imaging based on the boot config's root_device
	var imageDisk *linodego.InstanceDisk
	bootDiskID, ok := diskLabelToID[bootDiskLabel]
	if !ok {
		return handleError("Failed to find boot disk", fmt.Errorf("disk with label %q not found", bootDiskLabel))
	}

	disks, err := s.client.ListInstanceDisks(ctx, instance.ID, nil)
	if err != nil {
		return handleError("Failed to list instance disks", err)
	}

	for _, disk := range disks {
		if disk.ID == bootDiskID {
			imageDisk = &disk
			break
		}
	}

	if imageDisk == nil {
		return handleError("Failed to find boot disk", fmt.Errorf("disk with ID %d not found", bootDiskID))
	}

	state.Put("disk", imageDisk)

	// Boot the instance with the first configuration profile
	if bootConfigID != 0 {
		ui.Say(fmt.Sprintf("Booting Linode with config ID %d...", bootConfigID))
		err = s.client.BootInstance(ctx, instance.ID, bootConfigID)
		if err != nil {
			return handleError("Failed to boot Linode", err)
		}

		// Wait for instance to be running
		instance, err = s.client.WaitForInstanceStatus(ctx, instance.ID, linodego.InstanceRunning, int(c.StateTimeout.Seconds()))
		if err != nil {
			return handleError("Failed to wait for Linode to be running", err)
		}
		state.Put("instance", instance)
		ui.Say("Linode is now running")
	}

	return multistep.ActionContinue
}

func (s *stepCreateDiskConfig) Cleanup(state multistep.StateBag) {
	// Disks and configs are deleted when the instance is deleted
	// No additional cleanup needed here
}
