//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type LinodeInterface,InterfaceDefaultRoute,PublicInterface,PublicInterfaceIPv4,PublicInterfaceIPv6,PublicInterfaceIPv4Address,PublicInterfaceIPv6Range,VPCInterface,VPCInterfaceIPv4,VPCInterfaceIPv4Address,VPCInterfaceIPv4Range,VLANInterface
package linode

type LinodeInterface struct {
	// The enabled firewall to secure a VPC or public interface. Not allowed for VLAN interfaces.
	FirewallID *int `mapstructure:"firewall_id" required:"false"`

	// Indicates if the interface serves as the default route when multiple interfaces are
	// eligible for this role.
	DefaultRoute *InterfaceDefaultRoute `mapstructure:"default_route" required:"false"`

	// Public interface settings. A Linode can have only one public interface.
	// A public interface can have both IPv4 and IPv6 configurations.
	Public *PublicInterface `mapstructure:"public" required:"false"`

	// VPC interface settings.
	VPC *VPCInterface `mapstructure:"vpc" required:"false"`

	// VLAN interface settings.
	VLAN *VLANInterface `mapstructure:"vlan" required:"false"`
}

type InterfaceDefaultRoute struct {
	// Whether this interface is used for the IPv4 default route.
	IPv4 *bool `mapstructure:"ipv4" required:"false"`

	// Whether this interface is used for the IPv6 default route.
	IPv6 *bool `mapstructure:"ipv6" required:"false"`
}

type PublicInterface struct {
	// IPv4 address settings for this public interface. If omitted,
	// a public IPv4 address is automatically allocated.
	IPv4 *PublicInterfaceIPv4 `mapstructure:"ipv4" required:"false"`

	// IPv6 address settings for the public interface.
	IPv6 *PublicInterfaceIPv6 `mapstructure:"ipv6" required:"false"`
}

type PublicInterfaceIPv4 struct {
	// Blocks of IPv4 addresses to assign to this interface. Setting any to auto
	// allocates a public IPv4 address.
	Addresses []PublicInterfaceIPv4Address `mapstructure:"address" required:"false"`
}

type PublicInterfaceIPv4Address struct {
	// The interface's public IPv4 address. You can specify which public IPv4
	// address to configure for the interface. Setting this to auto automatically
	// allocates a public address.
	Address string `mapstructure:"address" required:"true"`

	// The IPv4 primary address configures the source address for routes within
	// the Linode on the corresponding network interface.
	//
	// - Don't set this to false if there's only one address in the addresses array.
	// - If more than one address is provided, primary can be set to true for one address.
	// - If only one address is present in the addresses array, this address is automatically set as the primary address.
	Primary *bool `mapstructure:"primary" required:"false"`
}

type PublicInterfaceIPv6 struct {
	// IPv6 address ranges to assign to this interface. If omitted, no ranges are assigned.
	Ranges []PublicInterfaceIPv6Range `mapstructure:"ranges" required:"false"`
}

type PublicInterfaceIPv6Range struct {
	// Your assigned IPv6 range in CIDR notation (2001:0db8::1/64) or prefix (/64).
	//
	// - The prefix of /64 or /56 block of IPv6 addresses.
	// - If provided in CIDR notation, the prefix must be within the assigned ranges for the Linode.
	Range string `mapstructure:"range" required:"true"`
}

type VPCInterface struct {
	// The VPC subnet identifier for this interface. Your subnet’s VPC must be in
	// the same data center (region) as the Linode.
	SubnetID int `mapstructure:"subnet_id" required:"true"`

	// Interfaces can be configured with IPv4 addresses or ranges.
	IPv4 *VPCInterfaceIPv4 `mapstructure:"ipv4" required:"false"`
}
type VPCInterfaceIPv4 struct {
	// IPv4 address settings for this VPC interface.
	Addresses []VPCInterfaceIPv4Address `mapstructure:"addresses" required:"false"`

	// VPC IPv4 ranges.
	Ranges []VPCInterfaceIPv4Range `mapstructure:"ranges" required:"false"`
}

type VPCInterfaceIPv4Address struct {
	// Specifies which IPv4 address to use in the VPC subnet. You can specify which
	// VPC Ipv4 address in the subnet to configure for the interface. You can't use
	// an IPv4 address taken from another Linode or interface, or the first two or
	// last two addresses in the VPC subnet. When address is set to `auto`, an IP
	// address from the subnet is automatically assigned.
	Address string `mapstructure:"address" required:"true"`

	// The IPv4 primary address is used to configure the source address for routes
	// within the Linode on the corresponding network interface.
	Primary *bool `mapstructure:"primary" required:"false"`

	// The 1:1 NAT IPv4 address used to associate a public IPv4 address with the
	// interface's VPC subnet IPv4 address.
	NAT1To1Address *string `mapstructure:"nat_1_1_address" required:"false"`
}

type VPCInterfaceIPv4Range struct {
	// VPC IPv4 ranges.
	Range string `mapstructure:"range" required:"true"`
}

type VLANInterface struct {
	// The VLAN's unique label. VLAN interfaces on the same Linode must have a unique `vlan_label`.
	VLANLabel string `mapstructure:"vlan_label" required:"true"`

	// This VLAN interface's private IPv4 address in classless inter-domain routing (CIDR) notation.
	IPAMAddress *string `mapstructure:"ipam_address" required:"false"`
}
