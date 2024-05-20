//go:generate packer-sdc struct-markdown
package helper

// The common configuration options related to Linode services
type LinodeCommon struct {
	// The Linode API token required for provision Linode resources.
	// This can also be specified in `LINODE_TOKEN` environment variable.
	// Saving the token in the environment or centralized vaults
	// can reduce the risk of the token being leaked from the codebase.
	// `images:read_write`, `linodes:read_write`, and `events:read_only`
	// scopes are required for the API token.
	PersonalAccessToken string `mapstructure:"linode_token"`
}
