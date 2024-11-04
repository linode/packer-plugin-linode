Type: `linode-image`

The Linode Image data source matches or filters the ID or label of both public images on
Linode and private images in your account using regular expression (regex) or an exact
match.

You can get the latest list of available public images on Linode via the
[Linode Image List API](https://techdocs.akamai.com/linode-api/reference/get-images).

## Examples

```hcl
data "linode-image" "latest_ubuntu" {
    id_regex = "linode/ubuntu.*"
    latest = true
}

source "linode" "example" {
  image             = data.linode-image.latest_ubuntu.id
  image_description = "My Private Image"
  image_label       = "my-packaer-private-linode-image-test"
  instance_label    = "temporary-linode-image"
  instance_type     = "g6-nanode-1"
  region            = "us-mia"
  ssh_username      = "root"
}

build {
  sources = ["source.linode.example"]
}
```

```hcl
data "linode-image" "latest_ubuntu_lts" {
    label_regex = "Ubuntu [0-9]+\\.[0-9]+ LTS"
    latest = true
}
```

```hcl
data "linode-image" "ubuntu22_lts" {
    id = "linode/ubuntu22.04"
    latest = true
}
```

## Configuration Reference:

<!-- Code generated from the comments of the Config struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `label` (string) - Matching the label of an image by exact label

- `label_regex` (string) - Matching the label of an image by a regular expression

- `id` (string) - Matching the ID of an image by exact ID

- `id_regex` (string) - Matching the ID of an image by a regular expression

- `latest` (bool) - Whether to use the latest created image when there are multiple matches

<!-- End of code generated from the comments of the Config struct in datasource/image/data.go; -->

<!-- Code generated from the comments of the LinodeCommon struct in helper/common.go; DO NOT EDIT MANUALLY -->

- `linode_token` (string) - The Linode API token required for provision Linode resources.
  This can also be specified in `LINODE_TOKEN` environment variable.
  Saving the token in the environment or centralized vaults
  can reduce the risk of the token being leaked from the codebase.
  `images:read_write`, `linodes:read_write`, and `events:read_only`
  scopes are required for the API token.

- `api_ca_path` (string) - The path to a CA file to trust when making API requests.
  It can also be specified using the `LINODE_CA` environment variable.

<!-- End of code generated from the comments of the LinodeCommon struct in helper/common.go; -->


## Output:

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `id` (string) - The unique ID of this Image.

- `capabilities` ([]string) - A list containing the following possible capabilities of this Image:
  - cloud-init: This Image supports cloud-init with Metadata. Only applies to public Images.

- `created` (string) - When this Image was created.

- `created_by` (string) - The name of the User who created this Image, or “linode” for public Images.

- `deprecated` (bool) - Whether or not this Image is deprecated. Will only be true for deprecated public Images.

- `description` (string) - A detailed description of this Image.

- `eol` (string) - The date of the public Image’s planned end of life. `null` for private Images.

- `expiry` (string) - Expiry date of the image.
  Only Images created automatically from a deleted Linode (type=automatic) will expire.

- `is_public` (bool) - True if the Image is a public distribution image.
  False if Image is private Account-specific Image.

- `label` (string) - A short description of the Image.

- `size` (int) - The minimum size this Image needs to deploy. Size is in MB.

- `type` (string) - Enum: `manual` `automatic`
  How the Image was created.
  "Manual" Images can be created at any time.
  "Automatic" Images are created automatically from a deleted Linode.

- `updated` (string) - When this Image was last updated.

- `vendor` (string) - The upstream distribution vendor. `null` for private Images.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; -->
