packer {
  required_plugins {
    linode = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/linode"
    }
  }
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "linode" "example" {
  linode_token      = "Your Personal Access Token"
  image             = "linode/debian9"
  image_description = "My Private Image"
  image_label       = "private-image-${local.timestamp}"
  instance_label    = "temporary-linode-${local.timestamp}"
  instance_type     = "g6-nanode-1"
  region            = "us-east"
  ssh_username      = "root"
}

build {
  hcp_packer_registry {
    bucket_name = "linode-hcp-test"
    description = "A nice test description"
    bucket_labels = {
      "foo" = "bar"
    }
  }
  sources = ["source.linode.example"]
}