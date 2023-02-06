# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    linode = {
      version = ">= 1.0.1"
      source  = "github.com/hashicorp/linode"
    }
  }
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "linode" "example" {
  image             = "linode/debian9"
  image_description = "My Private Image"
  image_label       = "private-image-${local.timestamp}"
  instance_label    = "temporary-linode-${local.timestamp}"
  instance_type     = "g6-nanode-1"
  linode_token      = "YOUR API TOKEN"
  region            = "us-east"
  ssh_username      = "root"
}

build {
  sources = ["source.linode.example"]
}
