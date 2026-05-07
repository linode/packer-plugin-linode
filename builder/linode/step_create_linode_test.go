package linode

import (
	"testing"

	"github.com/linode/linodego"
)

func TestFlattenVPCInterface_IPv4AddressFields(t *testing.T) {
	vpc := &VPCInterface{
		SubnetID: 12345,
		IPv4: &VPCInterfaceIPv4{
			Addresses: []VPCInterfaceIPv4Address{
				{
					Address:        linodego.Pointer("auto"),
					Primary:        linodego.Pointer(true),
					NAT1To1Address: linodego.Pointer("192.0.2.10"),
				},
			},
			Ranges: []VPCInterfaceIPv4Range{
				{Range: "10.0.0.0/28"},
			},
		},
	}

	got := flattenVPCInterface(vpc)
	if got == nil {
		t.Fatal("flattenVPCInterface() returned nil")
	}
	if got.SubnetID != 12345 {
		t.Fatalf("subnet_id = %d, want 12345", got.SubnetID)
	}
	if got.IPv4 == nil {
		t.Fatal("flattenVPCInterface().IPv4 returned nil")
	}
	if got.IPv4.Addresses == nil || len(*got.IPv4.Addresses) != 1 {
		t.Fatalf("flattenVPCInterface().IPv4.Addresses = %v, want one address", got.IPv4.Addresses)
	}

	addr := (*got.IPv4.Addresses)[0]
	if addr.Address == nil || *addr.Address != "auto" {
		t.Fatalf("address = %v, want auto", addr.Address)
	}
	if addr.Primary == nil || *addr.Primary != true {
		t.Fatalf("primary = %v, want true", addr.Primary)
	}
	if addr.NAT1To1Address == nil || *addr.NAT1To1Address != "192.0.2.10" {
		t.Fatalf("nat_1_1_address = %v, want 192.0.2.10", addr.NAT1To1Address)
	}
	if got.IPv4.Ranges == nil || len(*got.IPv4.Ranges) != 1 {
		t.Fatalf("ranges = %v, want one range", got.IPv4.Ranges)
	}
	if (*got.IPv4.Ranges)[0].Range != "10.0.0.0/28" {
		t.Fatalf("range = %q, want 10.0.0.0/28", (*got.IPv4.Ranges)[0].Range)
	}
}

func TestFlattenVPCInterface_IPv6Fields(t *testing.T) {
	vpc := &VPCInterface{
		SubnetID: 12345,
		IPv6: &VPCInterfaceIPv6{
			SLAAC: []VPCInterfaceIPv6SLAAC{
				{Range: "2600:3c03:e000:123::/64"},
			},
			Ranges: []VPCInterfaceIPv6Range{
				{Range: "2600:3c03:e000:123:1::/64"},
			},
			IsPublic: linodego.Pointer(true),
		},
	}

	got := flattenVPCInterface(vpc)
	if got == nil {
		t.Fatal("flattenVPCInterface() returned nil")
	}
	if got.IPv6 == nil {
		t.Fatal("flattenVPCInterface().IPv6 returned nil")
	}
	if got.IPv6.SLAAC == nil || len(*got.IPv6.SLAAC) != 1 {
		t.Fatalf("slaac = %v, want one slaac range", got.IPv6.SLAAC)
	}
	if (*got.IPv6.SLAAC)[0].Range != "2600:3c03:e000:123::/64" {
		t.Fatalf("slaac range = %q, want 2600:3c03:e000:123::/64", (*got.IPv6.SLAAC)[0].Range)
	}
	if got.IPv6.Ranges == nil || len(*got.IPv6.Ranges) != 1 {
		t.Fatalf("ranges = %v, want one ipv6 range", got.IPv6.Ranges)
	}
	if (*got.IPv6.Ranges)[0].Range != "2600:3c03:e000:123:1::/64" {
		t.Fatalf("range = %q, want 2600:3c03:e000:123:1::/64", (*got.IPv6.Ranges)[0].Range)
	}
	if got.IPv6.IsPublic == nil || !*got.IPv6.IsPublic {
		t.Fatalf("is_public = %v, want true", got.IPv6.IsPublic)
	}
}

func TestFlattenConfigInterface_AllFields(t *testing.T) {
	vpcIP := &InterfaceIPv4{VPC: "10.0.0.2", NAT1To1: linodego.Pointer("198.51.100.2")}
	iface := Interface{
		VLANInterfaceAttributes: VLANInterfaceAttributes{
			Label:       "eth1",
			IPAMAddress: "10.0.0.2/24",
		},
		VPCInterfaceAttributes: VPCInterfaceAttributes{
			SubnetID: linodego.Pointer(999),
			IPv4:     vpcIP,
			IPRanges: []string{"10.0.0.3/32"},
		},
		Purpose: "vpc",
		Primary: true,
	}

	got := flattenConfigInterface(iface)
	if got.IPAMAddress != iface.IPAMAddress {
		t.Fatalf("ipam_address = %q, want %q", got.IPAMAddress, iface.IPAMAddress)
	}
	if got.Label != iface.Label {
		t.Fatalf("label = %q, want %q", got.Label, iface.Label)
	}
	if got.Purpose != linodego.ConfigInterfacePurpose(iface.Purpose) {
		t.Fatalf("purpose = %q, want %q", got.Purpose, iface.Purpose)
	}
	if !got.Primary {
		t.Fatal("primary = false, want true")
	}
	if got.SubnetID == nil || *got.SubnetID != 999 {
		t.Fatalf("subnet_id = %v, want 999", got.SubnetID)
	}
	if got.IPv4 == nil || got.IPv4.VPC != "10.0.0.2" {
		t.Fatalf("ipv4 = %v, want VPC 10.0.0.2", got.IPv4)
	}
	if got.IPv4.NAT1To1 == nil || *got.IPv4.NAT1To1 != "198.51.100.2" {
		t.Fatalf("nat_1_1 = %v, want 198.51.100.2", got.IPv4.NAT1To1)
	}
	if len(got.IPRanges) != 1 || got.IPRanges[0] != "10.0.0.3/32" {
		t.Fatalf("ip_ranges = %v, want [10.0.0.3/32]", got.IPRanges)
	}
}

func TestFlattenPublicInterface_AllFields(t *testing.T) {
	public := &PublicInterface{
		IPv4: &PublicInterfaceIPv4{
			Addresses: []PublicInterfaceIPv4Address{{
				Address: linodego.Pointer("auto"),
				Primary: linodego.Pointer(true),
			}},
		},
		IPv6: &PublicInterfaceIPv6{
			Ranges: []PublicInterfaceIPv6Range{{Range: "/64"}},
		},
	}

	got := flattenPublicInterface(public)
	if got == nil || got.IPv4 == nil || got.IPv6 == nil {
		t.Fatalf("flattenPublicInterface() = %v, want non-nil ipv4 and ipv6", got)
	}
	if got.IPv4.Addresses == nil || len(*got.IPv4.Addresses) != 1 {
		t.Fatalf("ipv4 addresses = %v, want one address", got.IPv4.Addresses)
	}
	addr := (*got.IPv4.Addresses)[0]
	if addr.Address == nil || *addr.Address != "auto" {
		t.Fatalf("ipv4 address = %v, want auto", addr.Address)
	}
	if addr.Primary == nil || !*addr.Primary {
		t.Fatalf("ipv4 primary = %v, want true", addr.Primary)
	}
	if got.IPv6.Ranges == nil || len(*got.IPv6.Ranges) != 1 || (*got.IPv6.Ranges)[0].Range != "/64" {
		t.Fatalf("ipv6 ranges = %v, want [/64]", got.IPv6.Ranges)
	}
}

func TestFlattenLinodeInterface_AllFields(t *testing.T) {
	fw := 123
	li := LinodeInterface{
		FirewallID: &fw,
		DefaultRoute: &InterfaceDefaultRoute{
			IPv4: linodego.Pointer(true),
			IPv6: linodego.Pointer(false),
		},
		Public: &PublicInterface{
			IPv4: &PublicInterfaceIPv4{
				Addresses: []PublicInterfaceIPv4Address{{Address: linodego.Pointer("auto")}},
			},
		},
		VPC: &VPCInterface{SubnetID: 99},
		VLAN: &VLANInterface{
			VLANLabel:   "vlan-1",
			IPAMAddress: linodego.Pointer("10.0.0.1/24"),
		},
	}

	got := flattenLinodeInterface(li)
	if got.FirewallID == nil || *got.FirewallID != 123 {
		t.Fatalf("firewall_id = %v, want 123", got.FirewallID)
	}
	if got.DefaultRoute == nil || got.DefaultRoute.IPv4 == nil || !*got.DefaultRoute.IPv4 {
		t.Fatalf("default route ipv4 = %v, want true", got.DefaultRoute)
	}
	if got.DefaultRoute.IPv6 == nil || *got.DefaultRoute.IPv6 {
		t.Fatalf("default route ipv6 = %v, want false", got.DefaultRoute.IPv6)
	}
	if got.Public == nil {
		t.Fatalf("public = nil, want non-nil")
	}
	if got.Public.IPv4 == nil || got.Public.IPv4.Addresses == nil || len(*got.Public.IPv4.Addresses) != 1 {
		t.Fatalf("public ipv4 addresses = %v, want one address", got.Public)
	}
	if addr := (*got.Public.IPv4.Addresses)[0]; addr.Address == nil || *addr.Address != "auto" {
		t.Fatalf("public ipv4 address = %v, want auto", addr.Address)
	}
	if got.Public.IPv6 != nil {
		t.Fatalf("public ipv6 = %v, want nil", got.Public.IPv6)
	}
	if got.VPC == nil || got.VPC.SubnetID != 99 {
		t.Fatalf("vpc = %v, want subnet_id 99", got.VPC)
	}
	if got.VPC.IPv4 != nil {
		t.Fatalf("vpc ipv4 = %v, want nil", got.VPC.IPv4)
	}
	if got.VLAN == nil || got.VLAN.VLANLabel != "vlan-1" {
		t.Fatalf("vlan = %v, want label vlan-1", got.VLAN)
	}
	if got.VLAN.IPAMAddress == nil || *got.VLAN.IPAMAddress != "10.0.0.1/24" {
		t.Fatalf("vlan ipam_address = %v, want 10.0.0.1/24", got.VLAN.IPAMAddress)
	}
}

func TestFlattenConfigInterfaceIPv4(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := flattenConfigInterfaceIPv4(nil); got != nil {
			t.Fatalf("flattenConfigInterfaceIPv4(nil) = %v, want nil", got)
		}
	})

	t.Run("maps all fields", func(t *testing.T) {
		nat := "198.51.100.9"
		got := flattenConfigInterfaceIPv4(&InterfaceIPv4{
			VPC:     "10.10.10.2",
			NAT1To1: &nat,
		})
		if got == nil {
			t.Fatal("flattenConfigInterfaceIPv4() returned nil")
		}
		if got.VPC != "10.10.10.2" {
			t.Fatalf("VPC = %q, want 10.10.10.2", got.VPC)
		}
		if got.NAT1To1 == nil || *got.NAT1To1 != nat {
			t.Fatalf("NAT1To1 = %v, want %q", got.NAT1To1, nat)
		}
	})
}

func TestFlattenVLANInterface(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := flattenVLANInterface(nil); got != nil {
			t.Fatalf("flattenVLANInterface(nil) = %v, want nil", got)
		}
	})

	t.Run("with and without ipam", func(t *testing.T) {
		gotNoIPAM := flattenVLANInterface(&VLANInterface{VLANLabel: "vlan-a"})
		if gotNoIPAM == nil || gotNoIPAM.VLANLabel != "vlan-a" {
			t.Fatalf("flattenVLANInterface(no ipam) = %v, want label vlan-a", gotNoIPAM)
		}
		if gotNoIPAM.IPAMAddress != nil {
			t.Fatalf("IPAMAddress = %v, want nil", gotNoIPAM.IPAMAddress)
		}

		ipam := "10.0.0.2/24"
		gotWithIPAM := flattenVLANInterface(&VLANInterface{VLANLabel: "vlan-b", IPAMAddress: &ipam})
		if gotWithIPAM == nil || gotWithIPAM.IPAMAddress == nil || *gotWithIPAM.IPAMAddress != ipam {
			t.Fatalf("flattenVLANInterface(with ipam) = %v, want ipam %q", gotWithIPAM, ipam)
		}
	})
}

func TestFlattenMetadata(t *testing.T) {
	t.Run("empty userdata returns nil", func(t *testing.T) {
		if got := flattenMetadata(Metadata{}); got != nil {
			t.Fatalf("flattenMetadata(empty) = %v, want nil", got)
		}
	})

	t.Run("userdata is mapped", func(t *testing.T) {
		m := Metadata{UserData: "IyEvYmluL2Jhc2gK"}
		got := flattenMetadata(m)
		if got == nil || got.UserData != m.UserData {
			t.Fatalf("flattenMetadata() = %v, want UserData %q", got, m.UserData)
		}
	})
}
