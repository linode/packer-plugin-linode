package linode

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/linode/packer-plugin-linode/helper/acceptance"
)

func TestBuilderAcc_basic(t *testing.T) {
	if skip := acceptance.TestAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-builder-basic",
		Type:     "linode",
		Template: testBuilderAccBasic,
	})
}

const testBuilderAccBasic = `
source "linode" "example" {
	image             = "linode/arch"
	instance_type     = "g6-nanode-1"
	region            = "us-mia"
	ssh_username      = "root"
}

build {
	sources = ["source.linode.example"]
}
`

func TestBuilderAcc_customDisksAndConfig(t *testing.T) {
	if skip := acceptance.TestAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-builder-custom-disks-config",
		Type:     "linode",
		Template: testBuilderAccCustomDisksConfig,
	})
}

const testBuilderAccCustomDisksConfig = `
source "linode" "custom" {
	instance_type     = "g6-nanode-1"
	region            = "us-mia"
	ssh_username      = "root"
	interface_generation = "legacy_config"
	
	disk {
		label      = "boot"
		size       = 25000
		image      = "linode/arch"
		filesystem = "ext4"
	}
	
	disk {
		label      = "swap"
		size       = 512
		filesystem = "swap"
	}
	
	config {
		label       = "my-config"
		comments    = "Boot configuration"
		kernel      = "linode/grub2"
		root_device = "/dev/sda"
		run_level   = "default"
		
		devices {
			sda { disk_label = "boot" }
			sdb { disk_label = "swap" }
		}
		
		helpers {
			updatedb_disabled   = true
			distro              = true
			modules_dep         = true
			network             = true
			devtmpfs_automount  = true
		}
		
		interface {
			purpose = "public"
		}
	}
}

build {
	sources = ["source.linode.custom"]
}
`

func TestBuilderAcc_customDisksWithLinodeInterface(t *testing.T) {
	if skip := acceptance.TestAccPreCheck(t); skip == true {
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-linode-builder-custom-disks-linode-interface",
		Type:     "linode",
		Template: testBuilderAccCustomDisksWithLinodeInterface,
	})
}

const testBuilderAccCustomDisksWithLinodeInterface = `
source "linode" "custom_linode_interface" {
	instance_type     = "g6-nanode-1"
	region            = "us-mia"
	ssh_username      = "root"
	interface_generation = "linode"
	
	# Newer linode_interface blocks work with custom disks
	linode_interface {
		public {
			ipv4 {
				address {
					address = "auto"
					primary = true
				}
			}
		}
	}
	
	disk {
		label      = "boot"
		size       = 25000
		image      = "linode/arch"
		filesystem = "ext4"
	}
	
	disk {
		label      = "swap"
		size       = 512
		filesystem = "swap"
	}
	
	config {
		label       = "my-config"
		comments    = "Boot configuration with linode_interface"
		kernel      = "linode/grub2"
		root_device = "/dev/sda"
		run_level   = "default"
		
		devices {
			sda { disk_label = "boot" }
			sdb { disk_label = "swap" }
		}
		
		helpers {
			updatedb_disabled   = true
			distro              = true
			modules_dep         = true
			network             = true
			devtmpfs_automount  = true
		}
	}
}

build {
	sources = ["source.linode.custom_linode_interface"]
}
`
