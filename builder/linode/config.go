//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Interface,InterfaceIPv4,Metadata,Disk,InstanceConfig,InstanceConfigDevice,InstanceConfigDevices,InstanceConfigHelpers

package linode

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/linode/packer-plugin-linode/helper"
)

type InterfaceIPv4 struct {
	// The IPv4 address from the VPC subnet to use for this interface.
	VPC string `mapstructure:"vpc"`

	// The public IPv4 address assigned to this Linode to be 1:1 NATed with the VPC IPv4 address.
	NAT1To1 *string `mapstructure:"nat_1_1"`
}

type Metadata struct {
	// Base64-encoded (cloud-config)[https://www.linode.com/docs/products/compute/compute-instances/guides/metadata-cloud-config/] data.
	UserData string `mapstructure:"user_data"`
}

// Disk represents a disk to be created for the Linode instance.
// See https://techdocs.akamai.com/linode-api/reference/post-add-linode-disk
type Disk struct {
	// The label for this disk.
	Label string `mapstructure:"label" required:"true"`

	// The size of the disk in MB. NOTE: Resizing a disk can only be done
	// when the Linode is offline and may take some time.
	Size int `mapstructure:"size" required:"true"`

	// An Image ID to deploy the Linode Disk from. If provided,
	// at least one of root_pass, authorized_keys, or authorized_users
	// must be provided to ensure access.
	Image string `mapstructure:"image" required:"false"`

	// The filesystem for the disk. Valid values are raw, swap, ext3, ext4, initrd.
	// Defaults to ext4.
	Filesystem string `mapstructure:"filesystem" required:"false"`

	// The root password for this disk when deploying from an image.
	RootPass string `mapstructure:"root_pass" required:"false"`

	// A list of public SSH keys to be installed on the disk as the root user's
	// ~/.ssh/authorized_keys file.
	AuthorizedKeys []string `mapstructure:"authorized_keys" required:"false"`

	// A list of usernames that will have their SSH keys installed as the root
	// user's ~/.ssh/authorized_keys file.
	AuthorizedUsers []string `mapstructure:"authorized_users" required:"false"`

	// A StackScript ID to deploy to this disk. Only applies to Image-based disks.
	StackscriptID int `mapstructure:"stackscript_id" required:"false"`

	// UDF data to pass to the StackScript.
	StackscriptData map[string]string `mapstructure:"stackscript_data" required:"false"`
}

// InstanceConfigDevice represents a device slot in a configuration profile.
type InstanceConfigDevice struct {
	// The label of the disk to assign to this device slot.
	// This will be resolved to the disk ID after disks are created.
	DiskLabel string `mapstructure:"disk_label" required:"false"`

	// The ID of the volume to assign to this device slot.
	VolumeID int `mapstructure:"volume_id" required:"false"`
}

// InstanceConfigDevices represents the device mappings for a configuration profile.
// Each device slot can contain either a disk or a volume.
type InstanceConfigDevices struct {
	// Device assignments for slots sda through sdz.
	SDA *InstanceConfigDevice `mapstructure:"sda" required:"false"`
	SDB *InstanceConfigDevice `mapstructure:"sdb" required:"false"`
	SDC *InstanceConfigDevice `mapstructure:"sdc" required:"false"`
	SDD *InstanceConfigDevice `mapstructure:"sdd" required:"false"`
	SDE *InstanceConfigDevice `mapstructure:"sde" required:"false"`
	SDF *InstanceConfigDevice `mapstructure:"sdf" required:"false"`
	SDG *InstanceConfigDevice `mapstructure:"sdg" required:"false"`
	SDH *InstanceConfigDevice `mapstructure:"sdh" required:"false"`
	SDI *InstanceConfigDevice `mapstructure:"sdi" required:"false"`
	SDJ *InstanceConfigDevice `mapstructure:"sdj" required:"false"`
	SDK *InstanceConfigDevice `mapstructure:"sdk" required:"false"`
	SDL *InstanceConfigDevice `mapstructure:"sdl" required:"false"`
	SDM *InstanceConfigDevice `mapstructure:"sdm" required:"false"`
	SDN *InstanceConfigDevice `mapstructure:"sdn" required:"false"`
	SDO *InstanceConfigDevice `mapstructure:"sdo" required:"false"`
	SDP *InstanceConfigDevice `mapstructure:"sdp" required:"false"`
	SDQ *InstanceConfigDevice `mapstructure:"sdq" required:"false"`
	SDR *InstanceConfigDevice `mapstructure:"sdr" required:"false"`
	SDS *InstanceConfigDevice `mapstructure:"sds" required:"false"`
	SDT *InstanceConfigDevice `mapstructure:"sdt" required:"false"`
	SDU *InstanceConfigDevice `mapstructure:"sdu" required:"false"`
	SDV *InstanceConfigDevice `mapstructure:"sdv" required:"false"`
	SDW *InstanceConfigDevice `mapstructure:"sdw" required:"false"`
	SDX *InstanceConfigDevice `mapstructure:"sdx" required:"false"`
	SDY *InstanceConfigDevice `mapstructure:"sdy" required:"false"`
	SDZ *InstanceConfigDevice `mapstructure:"sdz" required:"false"`

	// Device assignments for slots sdaa through sdaz.
	SDAA *InstanceConfigDevice `mapstructure:"sdaa" required:"false"`
	SDAB *InstanceConfigDevice `mapstructure:"sdab" required:"false"`
	SDAC *InstanceConfigDevice `mapstructure:"sdac" required:"false"`
	SDAD *InstanceConfigDevice `mapstructure:"sdad" required:"false"`
	SDAE *InstanceConfigDevice `mapstructure:"sdae" required:"false"`
	SDAF *InstanceConfigDevice `mapstructure:"sdaf" required:"false"`
	SDAG *InstanceConfigDevice `mapstructure:"sdag" required:"false"`
	SDAH *InstanceConfigDevice `mapstructure:"sdah" required:"false"`
	SDAI *InstanceConfigDevice `mapstructure:"sdai" required:"false"`
	SDAJ *InstanceConfigDevice `mapstructure:"sdaj" required:"false"`
	SDAK *InstanceConfigDevice `mapstructure:"sdak" required:"false"`
	SDAL *InstanceConfigDevice `mapstructure:"sdal" required:"false"`
	SDAM *InstanceConfigDevice `mapstructure:"sdam" required:"false"`
	SDAN *InstanceConfigDevice `mapstructure:"sdan" required:"false"`
	SDAO *InstanceConfigDevice `mapstructure:"sdao" required:"false"`
	SDAP *InstanceConfigDevice `mapstructure:"sdap" required:"false"`
	SDAQ *InstanceConfigDevice `mapstructure:"sdaq" required:"false"`
	SDAR *InstanceConfigDevice `mapstructure:"sdar" required:"false"`
	SDAS *InstanceConfigDevice `mapstructure:"sdas" required:"false"`
	SDAT *InstanceConfigDevice `mapstructure:"sdat" required:"false"`
	SDAU *InstanceConfigDevice `mapstructure:"sdau" required:"false"`
	SDAV *InstanceConfigDevice `mapstructure:"sdav" required:"false"`
	SDAW *InstanceConfigDevice `mapstructure:"sdaw" required:"false"`
	SDAX *InstanceConfigDevice `mapstructure:"sdax" required:"false"`
	SDAY *InstanceConfigDevice `mapstructure:"sday" required:"false"`
	SDAZ *InstanceConfigDevice `mapstructure:"sdaz" required:"false"`

	// Device assignments for slots sdba through sdbl.
	SDBA *InstanceConfigDevice `mapstructure:"sdba" required:"false"`
	SDBB *InstanceConfigDevice `mapstructure:"sdbb" required:"false"`
	SDBC *InstanceConfigDevice `mapstructure:"sdbc" required:"false"`
	SDBD *InstanceConfigDevice `mapstructure:"sdbd" required:"false"`
	SDBE *InstanceConfigDevice `mapstructure:"sdbe" required:"false"`
	SDBF *InstanceConfigDevice `mapstructure:"sdbf" required:"false"`
	SDBG *InstanceConfigDevice `mapstructure:"sdbg" required:"false"`
	SDBH *InstanceConfigDevice `mapstructure:"sdbh" required:"false"`
	SDBI *InstanceConfigDevice `mapstructure:"sdbi" required:"false"`
	SDBJ *InstanceConfigDevice `mapstructure:"sdbj" required:"false"`
	SDBK *InstanceConfigDevice `mapstructure:"sdbk" required:"false"`
	SDBL *InstanceConfigDevice `mapstructure:"sdbl" required:"false"`
}

// InstanceConfigHelpers are helper options that control Linux distribution specific tweaks.
type InstanceConfigHelpers struct {
	// Disables updatedb cron job to avoid disk thrashing.
	UpdateDBDisabled *bool `mapstructure:"updatedb_disabled" required:"false"`

	// Enables the Distro filesystem helper.
	Distro *bool `mapstructure:"distro" required:"false"`

	// Creates a modules dependency file for the Kernel.
	ModulesDep *bool `mapstructure:"modules_dep" required:"false"`

	// Configures network services.
	Network *bool `mapstructure:"network" required:"false"`

	// Automatically mounts devtmpfs.
	DevTmpFsAutomount *bool `mapstructure:"devtmpfs_automount" required:"false"`
}

// InstanceConfig represents a configuration profile for the Linode instance.
// See https://techdocs.akamai.com/linode-api/reference/post-add-linode-config
type InstanceConfig struct {
	// The label for this configuration profile.
	Label string `mapstructure:"label" required:"true"`

	// Whether to boot the Linode with this configuration profile.
	// Only one configuration profile can have this set to true.
	// If not specified, the first configuration profile will be used for booting.
	Booted bool `mapstructure:"booted" required:"false"`

	// Optional comments about this configuration profile.
	Comments string `mapstructure:"comments" required:"false"`

	// Device assignments for this configuration profile.
	Devices *InstanceConfigDevices `mapstructure:"devices" required:"true"`

	// Helper options for this configuration profile.
	Helpers *InstanceConfigHelpers `mapstructure:"helpers" required:"false"`

	// Legacy config interfaces for this configuration profile.
	// Conflicts with the top-level interface and linode_interface blocks.
	Interfaces []Interface `mapstructure:"interface" required:"false"`

	// Limits the amount of RAM the Linode can use. 0 (default) means no limit.
	MemoryLimit int `mapstructure:"memory_limit" required:"false"`

	// The kernel to boot with. Use "linode/latest-64bit" or "linode/grub2".
	// See https://api.linode.com/v4/linode/kernels for available kernels.
	Kernel string `mapstructure:"kernel" required:"false"`

	// The init RAM disk to use. This is optional and typically not needed.
	InitRD int `mapstructure:"init_rd" required:"false"`

	// The root device to boot from, e.g., "/dev/sda". When using custom disks,
	// the disk at this device slot in the booted configuration profile will be imaged.
	RootDevice string `mapstructure:"root_device" required:"false"`

	// The run level to boot into. Valid values are "default", "single", "binbash".
	RunLevel string `mapstructure:"run_level" required:"false"`

	// The virtualization mode. Valid values are "paravirt" or "fullvirt".
	VirtMode string `mapstructure:"virt_mode" required:"false"`
}

type VPCInterfaceAttributes struct {
	// The ID of the VPC Subnet this interface references.
	SubnetID *int `mapstructure:"subnet_id"`

	// The IPv4 configuration of this VPC interface.
	IPv4 *InterfaceIPv4 `mapstructure:"ipv4"`

	// The IPv4 ranges of this VPC interface.
	IPRanges []string `mapstructure:"ip_ranges"`
}

type VLANInterfaceAttributes struct {
	// The label of the VLAN this interface relates to.
	Label string `mapstructure:"label"`

	// This Network Interface’s private IP address in CIDR notation.
	IPAMAddress string `mapstructure:"ipam_address"`
}

type Interface struct {
	VLANInterfaceAttributes `mapstructure:",squash"`
	VPCInterfaceAttributes  `mapstructure:",squash"`

	// The purpose of this interface. (public, vlan, vpc)
	Purpose string `mapstructure:"purpose" required:"true"`

	// Whether this interface is a primary interface.
	Primary bool `mapstructure:"primary"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	helper.LinodeCommon `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	// Legacy Config Network Interfaces to add to this Linode’s Configuration Profile. Singular repeatable
	// block containing a `purpose`, a `label`, and an `ipam_address` field.
	Interfaces []Interface `mapstructure:"interface" required:"false"`

	// Newer Linode Network Interfaces to add to this Linode.
	LinodeInterfaces []LinodeInterface `mapstructure:"linode_interface" required:"false"`

	// The id of the region to launch the Linode instance in. Images are available in all
	// regions, but there will be less delay when deploying from the region where the image
	// was taken. See [regions](https://api.linode.com/v4/regions) for more information on
	// the available regions. Examples are `us-east`, `us-central`, `us-west`, `ap-south`,
	// `ca-east`, `ap-northeast`, `eu-central`, and `eu-west`.
	Region string `mapstructure:"region" required:"true"`

	// Public SSH keys need to be appended to the Linode instance.
	AuthorizedKeys []string `mapstructure:"authorized_keys" required:"false"`

	// Users whose SSH keys need to be appended to the Linode instance.
	AuthorizedUsers []string `mapstructure:"authorized_users" required:"false"`

	// The Linode type defines the pricing, CPU, disk, and RAM specs of the instance. See
	// [instance types](https://api.linode.com/v4/linode/types) for more information on the
	// available Linode instance types. Examples are `g6-nanode-1`, `g6-standard-2`,
	// `g6-highmem-16`, and `g6-dedicated-16`.
	InstanceType string `mapstructure:"instance_type" required:"true"`

	// The name assigned to the Linode Instance.
	Label string `mapstructure:"instance_label" required:"false"`

	// Tags to apply to the instance when it is created.
	Tags []string `mapstructure:"instance_tags" required:"false"`

	// An Image ID to deploy the Disk from. Official Linode Images start with `linode/`,
	// while user Images start with `private/`. See [images](https://api.linode.com/v4/images)
	// for more information on the Images available for use. Examples are `linode/debian12`,
	// `linode/debian13`, `linode/ubuntu24.04`, `linode/arch`, and `private/12345`.
	Image string `mapstructure:"image" required:"false"`

	// The disk size (MiB) allocated for swap space.
	SwapSize *int `mapstructure:"swap_size" required:"false"`

	// The size (MiB) of the primary boot disk. Any remaining disk space beyond
	// the boot disk and swap partition is left unallocated. If not specified,
	// the boot disk will use all available space after swap.
	BootSize *int `mapstructure:"boot_size" required:"false"`

	// The kernel to boot the instance with. This can be a kernel ID such as
	// "linode/latest-64bit" or "linode/grub2". See the available kernels at
	// https://api.linode.com/v4/linode/kernels.
	Kernel string `mapstructure:"kernel" required:"false"`

	// If true, the created Linode will have private networking enabled and assigned
	// a private IPv4 address.
	PrivateIP bool `mapstructure:"private_ip" required:"false"`

	// The root password of the Linode instance for building the image. Please note that when
	// you create a new Linode instance with an image, at least one of root_pass,
	// authorized_keys, or authorized_users must be provided
	RootPass string `mapstructure:"root_pass" required:"false"`

	// The name of the resulting image that will appear
	// in your account. Defaults to `packer-{{timestamp}}` (see [configuration
	// templates](/packer/docs/templates/legacy_json_templates/engine) for more info).
	ImageLabel string `mapstructure:"image_label" required:"false"`

	// The description of the resulting image that will appear in your account. Defaults to "".
	Description string `mapstructure:"image_description" required:"false"`

	// The time to wait, as a duration string, for the Linode instance to enter a desired state
	// (such as "running") before timing out. The default state timeout is "5m".
	StateTimeout time.Duration `mapstructure:"state_timeout" required:"false"`

	// This attribute is required only if the StackScript being deployed requires input data from
	// the User for successful completion. See User Defined Fields (UDFs) for more details.
	//
	// This attribute is required to be valid JSON.
	StackScriptData map[string]string `mapstructure:"stackscript_data" required:"false"`

	// A StackScript ID that will cause the referenced StackScript to be run during deployment
	// of this Linode. A compatible image is required to use a StackScript. To get a list of
	// available StackScript and their permitted Images see /stackscripts. This field cannot
	// be used when deploying from a Backup or a Private Image.
	StackScriptID int `mapstructure:"stackscript_id" required:"false"`

	// The time to wait, as a duration string, for the disk image to be created successfully
	// before timing out. The default image creation timeout is "10m".
	ImageCreateTimeout time.Duration `mapstructure:"image_create_timeout" required:"false"`

	// Whether the newly created image supports cloud-init.
	CloudInit bool `mapstructure:"cloud_init" required:"false"`

	// An object containing user-defined data relevant to the creation of Linodes.
	Metadata Metadata `mapstructure:"metadata" required:"false"`

	// The ID of the Firewall to attach this Linode to upon creation.
	FirewallID int `mapstructure:"firewall_id" required:"false"`

	// The regions where the outcome image will be replicated to.
	ImageRegions []string `mapstructure:"image_regions" required:"false"`

	// Image Share Group IDs to add the newly created private image to
	// immediately after image creation.
	ImageShareGroupIDs []int `mapstructure:"image_share_group_ids" required:"false"`

	// Specifies the interface type for the Linode. The value can be either
	// `legacy_config` or `linode`. The default value is determined by the
	// `interfaces_for_new_linodes` setting in the account settings.
	InterfaceGeneration string `mapstructure:"interface_generation" required:"false"`

	// Custom disks to create for this Linode. When specified, you are responsible
	// for creating all disks including the boot disk. See the `disk` block
	// documentation for available options.
	Disks []Disk `mapstructure:"disk" required:"false"`

	// Custom configuration profiles to create for this Linode. When specified,
	// you are responsible for creating all configuration profiles.
	// See the `config` block documentation for available options.
	InstanceConfigs []InstanceConfig `mapstructure:"config" required:"false"`
}

// parseRootDevice extracts the device slot name from a root_device path.
// e.g., "/dev/sda" -> "sda", "/dev/sdab" -> "sdab"
func parseRootDevice(rootDevice string) string {
	rootDevice = strings.TrimSpace(rootDevice)
	if strings.HasPrefix(rootDevice, "/dev/") {
		return strings.TrimPrefix(rootDevice, "/dev/")
	}
	return rootDevice
}

// getBootConfig returns the configuration profile that will be booted.
// Returns the first config with booted=true, or the first config if none have booted=true.
// Returns nil if there are no configs.
func (c *Config) getBootConfig() *InstanceConfig {
	if len(c.InstanceConfigs) == 0 {
		return nil
	}

	for i := range c.InstanceConfigs {
		if c.InstanceConfigs[i].Booted {
			return &c.InstanceConfigs[i]
		}
	}
	return &c.InstanceConfigs[0]
}

// getDeviceAtSlot returns the device configuration at the given slot name.
// Returns nil if the slot is empty or not found.
func (d *InstanceConfigDevices) getDeviceAtSlot(slot string) *InstanceConfigDevice {
	if d == nil {
		return nil
	}

	slot = strings.ToLower(slot)
	switch slot {
	case "sda":
		return d.SDA
	case "sdb":
		return d.SDB
	case "sdc":
		return d.SDC
	case "sdd":
		return d.SDD
	case "sde":
		return d.SDE
	case "sdf":
		return d.SDF
	case "sdg":
		return d.SDG
	case "sdh":
		return d.SDH
	case "sdi":
		return d.SDI
	case "sdj":
		return d.SDJ
	case "sdk":
		return d.SDK
	case "sdl":
		return d.SDL
	case "sdm":
		return d.SDM
	case "sdn":
		return d.SDN
	case "sdo":
		return d.SDO
	case "sdp":
		return d.SDP
	case "sdq":
		return d.SDQ
	case "sdr":
		return d.SDR
	case "sds":
		return d.SDS
	case "sdt":
		return d.SDT
	case "sdu":
		return d.SDU
	case "sdv":
		return d.SDV
	case "sdw":
		return d.SDW
	case "sdx":
		return d.SDX
	case "sdy":
		return d.SDY
	case "sdz":
		return d.SDZ
	case "sdaa":
		return d.SDAA
	case "sdab":
		return d.SDAB
	case "sdac":
		return d.SDAC
	case "sdad":
		return d.SDAD
	case "sdae":
		return d.SDAE
	case "sdaf":
		return d.SDAF
	case "sdag":
		return d.SDAG
	case "sdah":
		return d.SDAH
	case "sdai":
		return d.SDAI
	case "sdaj":
		return d.SDAJ
	case "sdak":
		return d.SDAK
	case "sdal":
		return d.SDAL
	case "sdam":
		return d.SDAM
	case "sdan":
		return d.SDAN
	case "sdao":
		return d.SDAO
	case "sdap":
		return d.SDAP
	case "sdaq":
		return d.SDAQ
	case "sdar":
		return d.SDAR
	case "sdas":
		return d.SDAS
	case "sdat":
		return d.SDAT
	case "sdau":
		return d.SDAU
	case "sdav":
		return d.SDAV
	case "sdaw":
		return d.SDAW
	case "sdax":
		return d.SDAX
	case "sday":
		return d.SDAY
	case "sdaz":
		return d.SDAZ
	case "sdba":
		return d.SDBA
	case "sdbb":
		return d.SDBB
	case "sdbc":
		return d.SDBC
	case "sdbd":
		return d.SDBD
	case "sdbe":
		return d.SDBE
	case "sdbf":
		return d.SDBF
	case "sdbg":
		return d.SDBG
	case "sdbh":
		return d.SDBH
	case "sdbi":
		return d.SDBI
	case "sdbj":
		return d.SDBJ
	case "sdbk":
		return d.SDBK
	case "sdbl":
		return d.SDBL
	default:
		return nil
	}
}

// getBootDiskLabel returns the disk label of the boot disk (the disk that will be imaged).
// It derives this from the booted config's root_device setting.
// Returns an error if the root_device is not set, doesn't point to a valid device,
// or the device is a volume instead of a disk.
func (c *Config) getBootDiskLabel() (string, error) {
	bootConfig := c.getBootConfig()
	if bootConfig == nil {
		return "", errors.New("no configuration profile found")
	}

	if bootConfig.RootDevice == "" {
		return "", fmt.Errorf("root_device is required in the boot configuration profile %q when using custom disks", bootConfig.Label)
	}

	slot := parseRootDevice(bootConfig.RootDevice)
	device := bootConfig.Devices.getDeviceAtSlot(slot)
	if device == nil {
		return "", fmt.Errorf("root_device %q points to device slot %q which has no disk or volume assigned in config %q",
			bootConfig.RootDevice, slot, bootConfig.Label)
	}

	if device.DiskLabel == "" {
		if device.VolumeID != 0 {
			return "", fmt.Errorf("root_device %q points to a volume, not a disk; images can only be created from disks",
				bootConfig.RootDevice)
		}
		return "", fmt.Errorf("root_device %q points to device slot %q which has no disk_label set",
			bootConfig.RootDevice, slot)
	}

	return device.DiskLabel, nil
}

func (c *Config) Prepare(raws ...any) ([]string, error) {
	if err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...); err != nil {
		return nil, err
	}

	var errs *packersdk.MultiError

	// Defaults

	if c.PersonalAccessToken == "" {
		// Default to environment variable for linode_token, if it exists
		c.PersonalAccessToken = os.Getenv("LINODE_TOKEN")
	}

	if c.APICAPath == "" {
		c.APICAPath = os.Getenv("LINODE_CA")
	}

	if c.ImageLabel == "" {
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.ImageLabel = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render image name: %s", err))
		}
	}

	if c.Label == "" {
		// Default to packer-[time-ordered-uuid]
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.Label = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unable to render Linode label: %s", err))
		}
	}

	if c.StateTimeout == 0 {
		// Default to 5 minute timeouts waiting for state change
		c.StateTimeout = 5 * time.Minute
	}

	if c.ImageCreateTimeout == 0 {
		// Default to 10 minute timeouts waiting for image creation
		c.ImageCreateTimeout = 10 * time.Minute
	}

	if strings.TrimSpace(c.RootPass) != "" {
		c.Comm.SSHPassword = c.RootPass
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	// When creating a Linode from an image, at least one of root_pass, authorized_keys, or authorized_users
	// must be provided to ensure access to the instance
	if c.Image != "" {
		hasRootPass := strings.TrimSpace(c.RootPass) != ""
		hasKeys := len(c.AuthorizedKeys) > 0
		hasUsers := len(c.AuthorizedUsers) > 0

		if !hasRootPass && !hasKeys && !hasUsers {
			errs = packersdk.MultiErrorAppend(
				errs,
				fmt.Errorf(
					"when image is specified, at least one of root_pass, authorized_keys, or authorized_users must be provided",
				),
			)
		}
	}

	for _, d := range c.Disks {
		if strings.TrimSpace(d.Image) == "" {
			continue
		}

		hasRootPass := strings.TrimSpace(d.RootPass) != ""
		hasKeys := len(d.AuthorizedKeys) > 0
		hasUsers := len(d.AuthorizedUsers) > 0

		if !hasRootPass && !hasKeys && !hasUsers {
			errs = packersdk.MultiErrorAppend(
				errs,
				fmt.Errorf(
					"disk %q: when image is specified, at least one of root_pass, authorized_keys, or authorized_users must be provided",
					d.Label,
				),
			)
		}
	}

	bootLabel, err := c.getBootDiskLabel()
	if err == nil {
		for _, d := range c.Disks {
			if d.Label != bootLabel {
				continue
			}

			if strings.TrimSpace(d.Image) == "" {
				break
			}

			hasRootPass := strings.TrimSpace(d.RootPass) != ""
			hasKeys := len(d.AuthorizedKeys) > 0
			hasUsers := len(d.AuthorizedUsers) > 0

			if !hasRootPass && !hasKeys && !hasUsers {
				errs = packersdk.MultiErrorAppend(
					errs,
					fmt.Errorf(
						"boot disk %q must define root_pass, authorized_keys, or authorized_users",
						d.Label,
					),
				)
			}

			break
		}
	}

	if c.PersonalAccessToken == "" {
		// Required configurations that will display errors if not set
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("linode_token is required"))
	}

	if c.Region == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.InstanceType == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("instance_type is required"))
	}

	if c.Image == "" && len(c.Disks) == 0 {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("either image or custom disks must be specified"))
	}

	if c.Image != "" && len(c.Disks) > 0 {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("cannot specify both image and custom disks"))
	}

	// Disks and configs must be specified together - both or neither
	if (len(c.InstanceConfigs) == 0) != (len(c.Disks) == 0) {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("disk and config blocks must be specified together"))
	}

	// Set default root_device for configs that don't have one
	for i := range c.InstanceConfigs {
		if c.InstanceConfigs[i].RootDevice == "" {
			c.InstanceConfigs[i].RootDevice = "/dev/sda"
		}
	}

	if len(c.Disks) > 0 {
		// Validate disk labels are unique
		diskLabels := make(map[string]bool)
		for _, disk := range c.Disks {
			if disk.Label == "" {
				errs = packersdk.MultiErrorAppend(
					errs, errors.New("disk label cannot be empty"))
			} else if diskLabels[disk.Label] {
				errs = packersdk.MultiErrorAppend(
					errs, fmt.Errorf("duplicate disk label %q found", disk.Label))
			} else {
				diskLabels[disk.Label] = true
			}
		}

		// Validate that the boot config's root_device points to a valid disk
		bootDiskLabel, err := c.getBootDiskLabel()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
		} else {
			// Validate that the boot disk label exists in the disk blocks
			found := false
			for _, disk := range c.Disks {
				if disk.Label == bootDiskLabel {
					found = true
					break
				}
			}
			if !found {
				errs = packersdk.MultiErrorAppend(
					errs, fmt.Errorf("root_device points to disk %q which is not defined in the disk blocks", bootDiskLabel))
			}
		}

		if len(c.AuthorizedKeys) > 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("authorized_keys cannot be specified when using custom disks (specify in disk blocks instead)"))
		}

		if len(c.AuthorizedUsers) > 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("authorized_users cannot be specified when using custom disks (specify in disk blocks instead)"))
		}

		if c.SwapSize != nil {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("swap_size cannot be specified when using custom disks (create a swap disk instead)"))
		}

		if c.BootSize != nil {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("boot_size cannot be specified when using custom disks (specify size in disk blocks instead)"))
		}

		if c.Kernel != "" {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("kernel cannot be specified when using custom disks (specify in config blocks instead)"))
		}

		if c.StackScriptID > 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("stackscript_id cannot be specified when using custom disks (specify in disk blocks instead)"))
		}

		if len(c.StackScriptData) > 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("stackscript_data cannot be specified when using custom disks (specify in disk blocks instead)"))
		}

		if len(c.Interfaces) > 0 {
			errs = packersdk.MultiErrorAppend(
				errs, errors.New("interface blocks cannot be specified when using custom disks (specify in config blocks instead)"))
		}
	}

	if c.Tags == nil {
		c.Tags = make([]string, 0)
	}
	tagRe := regexp.MustCompile("^[[:print:]]{3,50}$")

	for _, t := range c.Tags {
		if !tagRe.MatchString(t) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid tag: %s", t))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packersdk.LogSecretFilter.Set(c.PersonalAccessToken)
	return nil, nil
}
