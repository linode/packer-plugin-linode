terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "1.26.1"
    }
    hcp = {
      source  = "hashicorp/hcp"
      version = "0.24.0"
    }
  }
}

// either use "token" below or set LINODE_TOKEN ENV VAR
provider "linode" {
  token = "YOUR LINODE TOKEN"
}

// either set the below values or
// set HCP_CLIENT_ID and HCP_CLIENT_SECRET
provider "hcp" {
  client_id = "YOUR HCP CLIENT ID"
  client_secret = "YOUR HCP CLIENT SECRET"
}

data "hcp_packer_iteration" "production_linode" {
  bucket_name = "linode-hcp-test"
  channel     = "production"
}

data "hcp_packer_image" "production_linode_image" {
  bucket_name    = "linode-hcp-test"
  cloud_provider = "linode"
  iteration_id   = data.hcp_packer_iteration.production_linode.ulid
  region         = "us-east"
}

resource "linode_instance" "production_linode_instance" {
  label           = "test-hcp-linode-instance"
  image           = data.hcp_packer_image.production_linode_image.cloud_image_id
  region          = "us-east"
  type            = "g6-nanode-1"
  root_pass       = "terr4form_test"
}