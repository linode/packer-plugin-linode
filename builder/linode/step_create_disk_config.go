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

func flattenDisk(d Disk, rootPass string) linodego.InstanceDiskCreateOptions {
	return linodego.InstanceDiskCreateOptions{
		Label:           d.Label,
		Size:            d.Size,
		Image:           d.Image,
		RootPass:        rootPass,
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
		return 0, nil
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

	// Return nil if neither disk nor volume is set
	if result.DiskID == 0 && result.VolumeID == 0 {
		return nil, nil
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

	// sda through sdz
	if result.SDA, err = flattenInstanceConfigDevice(d.SDA, diskLabelToID); err != nil {
		return result, fmt.Errorf("sda: %w", err)
	}
	if result.SDB, err = flattenInstanceConfigDevice(d.SDB, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdb: %w", err)
	}
	if result.SDC, err = flattenInstanceConfigDevice(d.SDC, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdc: %w", err)
	}
	if result.SDD, err = flattenInstanceConfigDevice(d.SDD, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdd: %w", err)
	}
	if result.SDE, err = flattenInstanceConfigDevice(d.SDE, diskLabelToID); err != nil {
		return result, fmt.Errorf("sde: %w", err)
	}
	if result.SDF, err = flattenInstanceConfigDevice(d.SDF, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdf: %w", err)
	}
	if result.SDG, err = flattenInstanceConfigDevice(d.SDG, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdg: %w", err)
	}
	if result.SDH, err = flattenInstanceConfigDevice(d.SDH, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdh: %w", err)
	}
	if result.SDI, err = flattenInstanceConfigDevice(d.SDI, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdi: %w", err)
	}
	if result.SDJ, err = flattenInstanceConfigDevice(d.SDJ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdj: %w", err)
	}
	if result.SDK, err = flattenInstanceConfigDevice(d.SDK, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdk: %w", err)
	}
	if result.SDL, err = flattenInstanceConfigDevice(d.SDL, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdl: %w", err)
	}
	if result.SDM, err = flattenInstanceConfigDevice(d.SDM, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdm: %w", err)
	}
	if result.SDN, err = flattenInstanceConfigDevice(d.SDN, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdn: %w", err)
	}
	if result.SDO, err = flattenInstanceConfigDevice(d.SDO, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdo: %w", err)
	}
	if result.SDP, err = flattenInstanceConfigDevice(d.SDP, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdp: %w", err)
	}
	if result.SDQ, err = flattenInstanceConfigDevice(d.SDQ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdq: %w", err)
	}
	if result.SDR, err = flattenInstanceConfigDevice(d.SDR, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdr: %w", err)
	}
	if result.SDS, err = flattenInstanceConfigDevice(d.SDS, diskLabelToID); err != nil {
		return result, fmt.Errorf("sds: %w", err)
	}
	if result.SDT, err = flattenInstanceConfigDevice(d.SDT, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdt: %w", err)
	}
	if result.SDU, err = flattenInstanceConfigDevice(d.SDU, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdu: %w", err)
	}
	if result.SDV, err = flattenInstanceConfigDevice(d.SDV, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdv: %w", err)
	}
	if result.SDW, err = flattenInstanceConfigDevice(d.SDW, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdw: %w", err)
	}
	if result.SDX, err = flattenInstanceConfigDevice(d.SDX, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdx: %w", err)
	}
	if result.SDY, err = flattenInstanceConfigDevice(d.SDY, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdy: %w", err)
	}
	if result.SDZ, err = flattenInstanceConfigDevice(d.SDZ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdz: %w", err)
	}

	// sdaa through sdaz
	if result.SDAA, err = flattenInstanceConfigDevice(d.SDAA, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaa: %w", err)
	}
	if result.SDAB, err = flattenInstanceConfigDevice(d.SDAB, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdab: %w", err)
	}
	if result.SDAC, err = flattenInstanceConfigDevice(d.SDAC, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdac: %w", err)
	}
	if result.SDAD, err = flattenInstanceConfigDevice(d.SDAD, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdad: %w", err)
	}
	if result.SDAE, err = flattenInstanceConfigDevice(d.SDAE, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdae: %w", err)
	}
	if result.SDAF, err = flattenInstanceConfigDevice(d.SDAF, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaf: %w", err)
	}
	if result.SDAG, err = flattenInstanceConfigDevice(d.SDAG, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdag: %w", err)
	}
	if result.SDAH, err = flattenInstanceConfigDevice(d.SDAH, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdah: %w", err)
	}
	if result.SDAI, err = flattenInstanceConfigDevice(d.SDAI, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdai: %w", err)
	}
	if result.SDAJ, err = flattenInstanceConfigDevice(d.SDAJ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaj: %w", err)
	}
	if result.SDAK, err = flattenInstanceConfigDevice(d.SDAK, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdak: %w", err)
	}
	if result.SDAL, err = flattenInstanceConfigDevice(d.SDAL, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdal: %w", err)
	}
	if result.SDAM, err = flattenInstanceConfigDevice(d.SDAM, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdam: %w", err)
	}
	if result.SDAN, err = flattenInstanceConfigDevice(d.SDAN, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdan: %w", err)
	}
	if result.SDAO, err = flattenInstanceConfigDevice(d.SDAO, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdao: %w", err)
	}
	if result.SDAP, err = flattenInstanceConfigDevice(d.SDAP, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdap: %w", err)
	}
	if result.SDAQ, err = flattenInstanceConfigDevice(d.SDAQ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaq: %w", err)
	}
	if result.SDAR, err = flattenInstanceConfigDevice(d.SDAR, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdar: %w", err)
	}
	if result.SDAS, err = flattenInstanceConfigDevice(d.SDAS, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdas: %w", err)
	}
	if result.SDAT, err = flattenInstanceConfigDevice(d.SDAT, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdat: %w", err)
	}
	if result.SDAU, err = flattenInstanceConfigDevice(d.SDAU, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdau: %w", err)
	}
	if result.SDAV, err = flattenInstanceConfigDevice(d.SDAV, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdav: %w", err)
	}
	if result.SDAW, err = flattenInstanceConfigDevice(d.SDAW, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaw: %w", err)
	}
	if result.SDAX, err = flattenInstanceConfigDevice(d.SDAX, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdax: %w", err)
	}
	if result.SDAY, err = flattenInstanceConfigDevice(d.SDAY, diskLabelToID); err != nil {
		return result, fmt.Errorf("sday: %w", err)
	}
	if result.SDAZ, err = flattenInstanceConfigDevice(d.SDAZ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdaz: %w", err)
	}

	// sdba through sdbl
	if result.SDBA, err = flattenInstanceConfigDevice(d.SDBA, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdba: %w", err)
	}
	if result.SDBB, err = flattenInstanceConfigDevice(d.SDBB, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbb: %w", err)
	}
	if result.SDBC, err = flattenInstanceConfigDevice(d.SDBC, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbc: %w", err)
	}
	if result.SDBD, err = flattenInstanceConfigDevice(d.SDBD, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbd: %w", err)
	}
	if result.SDBE, err = flattenInstanceConfigDevice(d.SDBE, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbe: %w", err)
	}
	if result.SDBF, err = flattenInstanceConfigDevice(d.SDBF, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbf: %w", err)
	}
	if result.SDBG, err = flattenInstanceConfigDevice(d.SDBG, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbg: %w", err)
	}
	if result.SDBH, err = flattenInstanceConfigDevice(d.SDBH, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbh: %w", err)
	}
	if result.SDBI, err = flattenInstanceConfigDevice(d.SDBI, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbi: %w", err)
	}
	if result.SDBJ, err = flattenInstanceConfigDevice(d.SDBJ, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbj: %w", err)
	}
	if result.SDBK, err = flattenInstanceConfigDevice(d.SDBK, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbk: %w", err)
	}
	if result.SDBL, err = flattenInstanceConfigDevice(d.SDBL, diskLabelToID); err != nil {
		return result, fmt.Errorf("sdbl: %w", err)
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

	// Create disks
	for _, diskCfg := range c.Disks {
		ui.Say(fmt.Sprintf("Creating disk: %s...", diskCfg.Label))

		diskOpts := flattenDisk(diskCfg, c.Comm.Password())

		// Always append SSH key from communicator config
		// Note: Top-level authorized_keys/authorized_users are validated to be empty when using custom disks
		if len(c.Comm.SSHPublicKey) > 0 {
			diskOpts.AuthorizedKeys = append(diskOpts.AuthorizedKeys, string(c.Comm.SSHPublicKey))
		}

		// Use the communicator password if root_pass not specified but image is
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

	// Validate that only one config has booted=true
	bootedCount := 0
	bootedIndex := -1
	for i, cfgProfile := range c.InstanceConfigs {
		if cfgProfile.Booted {
			bootedCount++
			bootedIndex = i
		}
	}

	if bootedCount > 1 {
		return handleError("Multiple configuration profiles marked as booted",
			errors.New("only one configuration profile can have 'booted' set to true"))
	}

	// Determine which config to boot: the one marked as booted, or the first one
	bootConfigIndex := 0
	if bootedIndex >= 0 {
		bootConfigIndex = bootedIndex
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

	// Find the boot disk (first non-swap disk) and store it for image creation
	var bootDisk *linodego.InstanceDisk
	disks, err := s.client.ListInstanceDisks(ctx, instance.ID, nil)
	if err != nil {
		return handleError("Failed to list instance disks", err)
	}

	for _, disk := range disks {
		if disk.Filesystem != linodego.FilesystemSwap {
			bootDisk = &disk
			break
		}
	}

	if bootDisk == nil {
		return handleError("Failed to find boot disk", errors.New("no suitable disk was found"))
	}

	state.Put("disk", bootDisk)

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
