//go:generate packer-sdc struct-markdown
package helper

type LinodeCommon struct {
	// The Linode API token. This can also be specified in LINODE_TOKEN environment variable
	PersonalAccessToken string `mapstructure:"linode_token"`
}
