//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Interface,InterfaceIPv4,Metadata

package linode

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
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

	// Network Interfaces to add to this Linode’s Configuration Profile. Singular repeatable
	// block containing a `purpose`, a `label`, and an `ipam_address` field.
	Interfaces []Interface `mapstructure:"interface" required:"false"`

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
	// for more information on the Images available for use. Examples are `linode/debian9`,
	// `linode/fedora28`, `linode/ubuntu18.04`, `linode/arch`, and `private/12345`.
	Image string `mapstructure:"image" required:"true"`

	// The disk size (MiB) allocated for swap space.
	SwapSize int `mapstructure:"swap_size" required:"false"`

	// If true, the created Linode will have private networking enabled and assigned
	// a private IPv4 address.
	PrivateIP bool `mapstructure:"private_ip" required:"false"`

	// The root password of the Linode instance for building the image. Please note that when
	// you create a new Linode instance with a private image, you will be required to setup a
	// new root password.
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
}

func createRandomRootPassword() (string, error) {
	rawRootPass := make([]byte, 50)
	_, err := rand.Read(rawRootPass)
	if err != nil {
		return "", fmt.Errorf("failed to generate random password")
	}
	rootPass := base64.StdEncoding.EncodeToString(rawRootPass)
	return rootPass, nil
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

	if c.ImageLabel == "" {
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.ImageLabel = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Unable to render image name: %s", err))
		}
	}

	if c.Label == "" {
		// Default to packer-[time-ordered-uuid]
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.Label = def
		} else {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Unable to render Linode label: %s", err))
		}
	}

	if c.RootPass == "" {
		var err error
		c.RootPass, err = createRandomRootPassword()
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Unable to generate root_pass: %s", err))
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

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}

	c.Comm.SSHPassword = c.RootPass

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

	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("image is required"))
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
