package image

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/linode/linodego"
	"github.com/linode/packer-plugin-linode/helper"
	"github.com/zclconf/go-cty/cty"
)

type Datasource struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	helper.LinodeCommon `mapstructure:",squash"`

	// Matching the label of an image by exact label
	Label string `mapstructure:"label"`

	// Matching the label of an image by a regular expression
	LabelRegex string `mapstructure:"label_regex"`

	// Matching the ID of an image by exact ID
	ID string `mapstructure:"id"`

	// Matching the ID of an image by a regular expression
	IDRegex string `mapstructure:"id_regex"`

	// Whether to use the latest created image when there are multiple matches
	Latest bool `mapstructure:"latest"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.PersonalAccessToken == "" {
		envToken := os.Getenv(helper.TokenEnvVar)
		if envToken == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf(
				"A Linode API token is required. You can specify it in an "+
					"environment variable %q or set linode_token "+
					"attribute in the datasource block.",
				helper.TokenEnvVar,
			))
		}
		d.config.PersonalAccessToken = envToken
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

type DatasourceOutput struct {
	// The unique ID of this Image.
	ID string `mapstructure:"id"`

	// A list containing the following possible capabilities of this Image:
	// - cloud-init: This Image supports cloud-init with Metadata. Only applies to public Images.
	Capabilities []string `mapstructure:"capabilities"`

	// When this Image was created.
	Created string `mapstructure:"created"`

	// The name of the User who created this Image, or “linode” for public Images.
	CreatedBy string `mapstructure:"created_by"`

	// Whether or not this Image is deprecated. Will only be true for deprecated public Images.
	Deprecated bool `mapstructure:"deprecated"`

	// A detailed description of this Image.
	Description string `mapstructure:"description"`

	// The date of the public Image’s planned end of life. `null` for private Images.
	EOL string `mapstructure:"eol"`

	// Expiry date of the image.
	// Only Images created automatically from a deleted Linode (type=automatic) will expire.
	Expiry string `mapstructure:"expiry"`

	// True if the Image is a public distribution image.
	// False if Image is private Account-specific Image.
	IsPublic bool `mapstructure:"is_public"`

	// A short description of the Image.
	Label string `mapstructure:"label"`

	// The minimum size this Image needs to deploy. Size is in MB.
	Size int `mapstructure:"size"`

	// Enum: `manual` `automatic`
	// How the Image was created.
	// "Manual" Images can be created at any time.
	// "Automatic" Images are created automatically from a deleted Linode.
	Type string `mapstructure:"type"`

	// When this Image was last updated.
	Updated string `mapstructure:"updated"`

	// The upstream distribution vendor. `null` for private Images.
	Vendor string `mapstructure:"vendor"`
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	var client *linodego.Client
	var err error

	if d.config.APICAPath != "" {
		client, err = helper.NewLinodeClientWithCA(d.config.PersonalAccessToken, d.config.APICAPath)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), err
		}
	} else {
		client = helper.NewLinodeClient(d.config.PersonalAccessToken)
	}

	filters := linodego.Filter{}

	// Label is API filterable
	if d.config.Label != "" {
		filters.AddField(linodego.Eq, "label", d.config.Label)
	}

	// we only want available images for the obvious reason
	filters.AddField(linodego.Eq, "status", "available")

	filterString, err := filters.MarshalJSON()
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	images, err := client.ListImages(
		context.Background(),
		linodego.NewListOptions(0, string(filterString)),
	)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	// filtering non-API filterable attributes
	image, err := filterImageResults(images, d.config)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	return hcl2helper.HCL2ValueFromConfig(getOutput(image), d.OutputSpec()), nil
}

func getOutput(image linodego.Image) DatasourceOutput {
	output := DatasourceOutput{
		ID:           image.ID,
		Capabilities: image.Capabilities,
		CreatedBy:    image.CreatedBy,
		Deprecated:   image.Deprecated,
		Description:  image.Description,
		IsPublic:     image.IsPublic,
		Label:        image.Label,
		Size:         image.Size,
		Type:         image.Type,
		Vendor:       image.Vendor,
		Created:      image.Created.Format(time.RFC3339),
		Updated:      image.Updated.Format(time.RFC3339),
	}

	if image.EOL != nil {
		output.EOL = image.EOL.Format(time.RFC3339)
	}

	if image.Expiry != nil {
		output.Expiry = image.Expiry.Format(time.RFC3339)
	}

	return output
}
