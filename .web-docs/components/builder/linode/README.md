Type: `linode`
Artifact BuilderId: `packer.linode`

The `linode` Packer builder is able to create [Linode
Images](https://www.linode.com/docs/platform/disk-images/linode-images/) for
use with [Linode](https://www.linode.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the
foundation of new servers that are launched within Linode.

The builder does _not_ manage images. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/packer/docs/templates/legacy_json_templates/communicator) can be configured for this
builder. In addition to the options defined there, a private key file
can also be supplied to override the typical auto-generated key:

- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


<!--
  Linode.com has DDOS protection that returns 403 for the markdown link checker
  so the domain has been added to the ignorepatterns in mlc_config.json

  See https://github.com/tcort/markdown-link-check/issues/109
-->

### Required

<!-- Code generated from the comments of the LinodeCommon struct in helper/common.go; DO NOT EDIT MANUALLY -->

- `linode_token` (string) - The Linode API token required for provision Linode resources.
  This can also be specified in `LINODE_TOKEN` environment variable.
  Saving the token in the environment or centralized vaults
  can reduce the risk of the token being leaked from the codebase.

<!-- End of code generated from the comments of the LinodeCommon struct in helper/common.go; -->

<!-- Code generated from the comments of the Config struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `region` (string) - The id of the region to launch the Linode instance in. Images are available in all
  regions, but there will be less delay when deploying from the region where the image
  was taken. See [regions](https://api.linode.com/v4/regions) for more information on
  the available regions. Examples are `us-east`, `us-central`, `us-west`, `ap-south`,
  `ca-east`, `ap-northeast`, `eu-central`, and `eu-west`.

- `instance_type` (string) - The Linode type defines the pricing, CPU, disk, and RAM specs of the instance. See
  [instance types](https://api.linode.com/v4/linode/types) for more information on the
  available Linode instance types. Examples are `g6-nanode-1`, `g6-standard-2`,
  `g6-highmem-16`, and `g6-dedicated-16`.

- `image` (string) - An Image ID to deploy the Disk from. Official Linode Images start with `linode/`,
  while user Images start with `private/`. See [images](https://api.linode.com/v4/images)
  for more information on the Images available for use. Examples are `linode/debian9`,
  `linode/fedora28`, `linode/ubuntu18.04`, `linode/arch`, and `private/12345`.

<!-- End of code generated from the comments of the Config struct in builder/linode/config.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `interface` ([]Interface) - Network Interfaces to add to this Linode’s Configuration Profile. Singular repeatable
  block containing a `purpose`, a `label`, and an `ipam_address` field.

- `authorized_keys` ([]string) - Public SSH keys need to be appended to the Linode instance.

- `authorized_users` ([]string) - Users whose SSH keys need to be appended to the Linode instance.

- `instance_label` (string) - The name assigned to the Linode Instance.

- `instance_tags` ([]string) - Tags to apply to the instance when it is created.

- `swap_size` (int) - The disk size (MiB) allocated for swap space.

- `private_ip` (bool) - If true, the created Linode will have private networking enabled and assigned
  a private IPv4 address.

- `root_pass` (string) - The root password of the Linode instance for building the image. Please note that when
  you create a new Linode instance with a private image, you will be required to setup a
  new root password.

- `image_label` (string) - The name of the resulting image that will appear
  in your account. Defaults to `packer-{{timestamp}}` (see [configuration
  templates](/packer/docs/templates/legacy_json_templates/engine) for more info).

- `image_description` (string) - The description of the resulting image that will appear in your account. Defaults to "".

- `state_timeout` (duration string | ex: "1h5m2s") - The time to wait, as a duration string, for the Linode instance to enter a desired state
  (such as "running") before timing out. The default state timeout is "5m".

- `stackscript_data` (map[string]string) - This attribute is required only if the StackScript being deployed requires input data from
  the User for successful completion. See User Defined Fields (UDFs) for more details.
  
  This attribute is required to be valid JSON.

- `stackscript_id` (int) - A StackScript ID that will cause the referenced StackScript to be run during deployment
  of this Linode. A compatible image is required to use a StackScript. To get a list of
  available StackScript and their permitted Images see /stackscripts. This field cannot
  be used when deploying from a Backup or a Private Image.

- `image_create_timeout` (duration string | ex: "1h5m2s") - The time to wait, as a duration string, for the disk image to be created successfully
  before timing out. The default image creation timeout is "10m".

- `cloud_init` (bool) - Whether the newly created image supports cloud-init.

- `metadata` (Metadata) - An object containing user-defined data relevant to the creation of Linodes.

- `firewall_id` (int) - The ID of the Firewall to attach this Linode to upon creation.

<!-- End of code generated from the comments of the Config struct in builder/linode/config.go; -->


#### Interface

This section outlines the fields configurable for a single interface object.

##### Required Interface Common Attributes

<!-- Code generated from the comments of the Interface struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `purpose` (string) - The purpose of this interface. (public, vlan, vpc)

<!-- End of code generated from the comments of the Interface struct in builder/linode/config.go; -->


##### Optional Interface Common Attributes

<!-- Code generated from the comments of the Interface struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `primary` (bool) - Whether this interface is a primary interface.

<!-- End of code generated from the comments of the Interface struct in builder/linode/config.go; -->


##### VLAN-specific Attributes

<!-- Code generated from the comments of the VLANInterfaceAttributes struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `label` (string) - The label of the VLAN this interface relates to.

- `ipam_address` (string) - This Network Interface’s private IP address in CIDR notation.

<!-- End of code generated from the comments of the VLANInterfaceAttributes struct in builder/linode/config.go; -->


##### VPC-specific Attributes

<!-- Code generated from the comments of the VPCInterfaceAttributes struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `subnet_id` (\*int) - The ID of the VPC Subnet this interface references.

- `ipv4` (\*InterfaceIPv4) - The IPv4 configuration of this VPC interface.

- `ip_ranges` ([]string) - The IPv4 ranges of this VPC interface.

<!-- End of code generated from the comments of the VPCInterfaceAttributes struct in builder/linode/config.go; -->


- `subnet_id` (int) - The ID of the VPC Subnet this interface references.

- `ipv4` (block) - The IPv4 configuration of this VPC interface.

###### VPC Interface IPv4 configuration object

<!-- Code generated from the comments of the InterfaceIPv4 struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `vpc` (string) - The IPv4 address from the VPC subnet to use for this interface.

- `nat_1_1` (\*string) - The public IPv4 address assigned to this Linode to be 1:1 NATed with the VPC IPv4 address.

<!-- End of code generated from the comments of the InterfaceIPv4 struct in builder/linode/config.go; -->


#### Metadata

This section outlines the fields configurable for a single metadata object.

<!-- Code generated from the comments of the Metadata struct in builder/linode/config.go; DO NOT EDIT MANUALLY -->

- `user_data` (string) - Base64-encoded (cloud-config)[https://www.linode.com/docs/products/compute/compute-instances/guides/metadata-cloud-config/] data.

<!-- End of code generated from the comments of the Metadata struct in builder/linode/config.go; -->


## Examples

### Basic Example

Here is a Linode builder example. The `linode_token` should be replaced with an
actual [Linode Personal Access
Token](https://www.linode.com/docs/platform/api/getting-started-with-the-linode-api/#get-an-access-token)
or in the config file or the environmental variable, `LINODE_TOKEN`.

**HCL2**

```hcl

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "linode" "example" {
  image             = "linode/debian11"
  image_description = "My Private Image"
  image_label       = "private-image-${local.timestamp}"
  instance_label    = "temporary-linode-${local.timestamp}"
  instance_type     = "g6-nanode-1"
  linode_token      = "YOUR API TOKEN"
  region            = "us-mia"
  ssh_username      = "root"
}

build {
  sources = ["source.linode.example"]
}

```

**JSON**

```json
{
  "source": {
    "linode": {
      "example": {
        "image": "linode/debian11",
        "linode_token": "YOUR API TOKEN",
        "region": "us-mia",
        "instance_type": "g6-nanode-1",
        "instance_label": "temporary-linode-{{timestamp}}",
        "image_label": "private-image-{{timestamp}}",
        "image_description": "My Private Image",
        "ssh_username": "root"
      }
    }
  },
  "build": {
    "sources": [
      "source.linode.example"
    ]
  }
}
```


### Complicated Example

**HCL2**

```hcl

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "linode" "example" {
  image             = "linode/debian11"
  image_description = "My Private Image"
  image_label       = "private-image-${local.timestamp}"
  instance_label    = "temporary-linode-${local.timestamp}"
  instance_type     = "g6-nanode-1"
  linode_token      = "YOUR API TOKEN"
  region            = "us-mia"
  ssh_username      = "root"
  private_ip        = true
  firewall_id       = 12345

  instance_tags     = ["abc", "foo=bar"]
  authorized_users  = ["your_authorized_username"]
  authorized_keys   = ["ssh-rsa AAAA_valid_public_ssh_key_123456785== user@their-computer"]
  stackscript_id    = 1177256
  stackscript_data  = {
    "test_key": "test_value_1"
  }

  interface {
    purpose       = "public"
  }

  interface {
    purpose       = "vpc"
    subnet_id     = 123
    ipv4 {
        vpc = "10.0.0.2"
        nat_1_1 = "any"
    }
  }

  metadata {
    user_data = base64encode(<<EOF
#cloud-config

write_files:
  - path: /root/helloworld.txt
    content: |
      Hello, world!
    owner: 'root:root'
    permissions: '0644'
EOF
)
  }
}

build {
  sources = ["source.linode.example"]
}

```

**JSON**

```json
{
  "source": {
    "linode": {
      "example": {
        "image": "linode/debian11",
        "region": "us-southeast",
        "instance_type": "g6-nanode-1",
        "instance_label": "temporary-linode-{{timestamp}}",
        "private_ip": true,
        "image_label": "private-image-{{timestamp}}",
        "image_description": "My Private Image",
        "ssh_username": "root",
        "authorized_users": [
          "your_authorized_username"
        ],
        "authorized_keys": [
          "ssh-rsa AAAA_valid_public_ssh_key_123456785== user@their-computer"
        ],
        "stackscript_id": 123,
        "stackscript_data": {
          "test_data": "test_value"
        },
        "interface": [
          {
            "purpose": "public",
            "label": "",
            "ipam_address": ""
          },
          {
            "purpose": "vlan",
            "label": "vlan-1",
            "ipam_address": "10.0.0.1/24"
          },
          {
            "purpose": "vlan",
            "label": "vlan-2",
            "ipam_address": "10.0.0.2/24"
          }
        ]
      }
    }
  },
  "build": {
    "sources": [
      "source.linode.example"
    ]
  }
}
```
