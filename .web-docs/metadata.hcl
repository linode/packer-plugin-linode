# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "Linode"
  description = "Create Linode Images for use with Linode."
  identifier = "packer/linode/linode"
  component {
    type = "builder"
    name = "Linode"
    slug = "linode"
  }
}
