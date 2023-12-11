//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Interface,InterfaceIPv4

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
)

type InterfaceIPv4 struct {
	VPC     string `mapstructure:"vpc"`
	NAT1To1 string `mapstructure:"nat_1_1"`
}

type Interface struct {
	Purpose     string         `mapstructure:"purpose"`
	Label       string         `mapstructure:"label"`
	IPAMAddress string         `mapstructure:"ipam_address"`
	Primary     bool           `mapstructure:"primary"`
	SubnetID    *int           `mapstructure:"subnet_id"`
	IPv4        *InterfaceIPv4 `mapstructure:"ipv4"`
	IPRanges    []string       `mapstructure:"ip_ranges"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	PersonalAccessToken string `mapstructure:"linode_token"`

	Interfaces         []Interface       `mapstructure:"interface"`
	Region             string            `mapstructure:"region"`
	AuthorizedKeys     []string          `mapstructure:"authorized_keys"`
	AuthorizedUsers    []string          `mapstructure:"authorized_users"`
	InstanceType       string            `mapstructure:"instance_type"`
	Label              string            `mapstructure:"instance_label"`
	Tags               []string          `mapstructure:"instance_tags"`
	Image              string            `mapstructure:"image"`
	SwapSize           int               `mapstructure:"swap_size"`
	PrivateIP          bool              `mapstructure:"private_ip"`
	RootPass           string            `mapstructure:"root_pass"`
	ImageLabel         string            `mapstructure:"image_label"`
	Description        string            `mapstructure:"image_description"`
	StateTimeout       time.Duration     `mapstructure:"state_timeout" required:"false"`
	StackScriptData    map[string]string `mapstructure:"stackscript_data"`
	StackScriptID      int               `mapstructure:"stackscript_id"`
	ImageCreateTimeout time.Duration     `mapstructure:"image_create_timeout" required:"false"`
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

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
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
	tagRe := regexp.MustCompile("^[[:alnum:]:_-]{1,255}$")

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
