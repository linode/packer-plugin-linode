package linode

import (
	"reflect"
	"strings"
	"testing"

	"github.com/linode/linodego"
)

func TestResolveDiskLabel(t *testing.T) {
	diskLabelToID := map[string]int{
		"boot": 12345,
		"swap": 67890,
		"data": 11111,
	}

	tests := []struct {
		name      string
		label     string
		wantID    int
		wantError bool
	}{
		{
			name:      "Valid label",
			label:     "boot",
			wantID:    12345,
			wantError: false,
		},
		{
			name:      "Another valid label",
			label:     "swap",
			wantID:    67890,
			wantError: false,
		},
		{
			name:      "Empty label - error",
			label:     "",
			wantID:    0,
			wantError: true,
		},
		{
			name:      "Nonexistent label",
			label:     "nonexistent",
			wantID:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := resolveDiskLabel(tt.label, diskLabelToID)

			if tt.wantError {
				if err == nil {
					t.Errorf("resolveDiskLabel() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("resolveDiskLabel() unexpected error: %v", err)
				}
			}

			if gotID != tt.wantID {
				t.Errorf("resolveDiskLabel() = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}

func TestFlattenInstanceConfigDevice(t *testing.T) {
	diskLabelToID := map[string]int{
		"boot": 12345,
		"swap": 67890,
	}

	tests := []struct {
		name      string
		device    *InstanceConfigDevice
		wantDisk  int
		wantVol   int
		wantNil   bool
		wantError bool
	}{
		{
			name:      "Nil device returns nil",
			device:    nil,
			wantNil:   true,
			wantError: false,
		},
		{
			name: "Valid disk label",
			device: &InstanceConfigDevice{
				DiskLabel: "boot",
			},
			wantDisk:  12345,
			wantError: false,
		},
		{
			name: "Valid volume ID",
			device: &InstanceConfigDevice{
				VolumeID: 99999,
			},
			wantVol:   99999,
			wantError: false,
		},
		{
			name: "Both disk and volume - error",
			device: &InstanceConfigDevice{
				DiskLabel: "boot",
				VolumeID:  99999,
			},
			wantError: true,
		},
		{
			name: "Invalid disk label",
			device: &InstanceConfigDevice{
				DiskLabel: "nonexistent",
			},
			wantError: true,
		},
		{
			name:      "Empty device - error",
			device:    &InstanceConfigDevice{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenInstanceConfigDevice(tt.device, diskLabelToID)

			if tt.wantError {
				if err == nil {
					t.Errorf("flattenInstanceConfigDevice() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("flattenInstanceConfigDevice() unexpected error: %v", err)
				return
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("flattenInstanceConfigDevice() expected nil but got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("flattenInstanceConfigDevice() unexpected nil result")
				return
			}

			if got.DiskID != tt.wantDisk {
				t.Errorf("flattenInstanceConfigDevice() DiskID = %v, want %v", got.DiskID, tt.wantDisk)
			}

			if got.VolumeID != tt.wantVol {
				t.Errorf("flattenInstanceConfigDevice() VolumeID = %v, want %v", got.VolumeID, tt.wantVol)
			}
		})
	}
}

func TestFlattenInstanceConfigDevices(t *testing.T) {
	diskLabelToID := map[string]int{
		"boot": 12345,
		"swap": 67890,
	}

	t.Run("Nil devices returns empty map", func(t *testing.T) {
		result, err := flattenInstanceConfigDevices(nil, diskLabelToID)
		if err != nil {
			t.Errorf("flattenInstanceConfigDevices() unexpected error: %v", err)
		}
		if result.SDA != nil || result.SDB != nil {
			t.Errorf("flattenInstanceConfigDevices() expected empty device slots")
		}
	})

	t.Run("Valid device mappings", func(t *testing.T) {
		devices := &InstanceConfigDevices{
			SDA: &InstanceConfigDevice{DiskLabel: "boot"},
			SDB: &InstanceConfigDevice{DiskLabel: "swap"},
		}

		result, err := flattenInstanceConfigDevices(devices, diskLabelToID)
		if err != nil {
			t.Errorf("flattenInstanceConfigDevices() unexpected error: %v", err)
		}

		if result.SDA == nil || result.SDA.DiskID != 12345 {
			t.Errorf("flattenInstanceConfigDevices() SDA = %v, want DiskID 12345", result.SDA)
		}

		if result.SDB == nil || result.SDB.DiskID != 67890 {
			t.Errorf("flattenInstanceConfigDevices() SDB = %v, want DiskID 67890", result.SDB)
		}
	})

	t.Run("Invalid disk label in device", func(t *testing.T) {
		devices := &InstanceConfigDevices{
			SDA: &InstanceConfigDevice{DiskLabel: "nonexistent"},
		}

		_, err := flattenInstanceConfigDevices(devices, diskLabelToID)
		if err == nil {
			t.Errorf("flattenInstanceConfigDevices() expected error for invalid disk label")
		}
	})
}

func TestFlattenDisk(t *testing.T) {
	tests := []struct {
		name string
		disk Disk
		want linodego.InstanceDiskCreateOptions
	}{
		{
			name: "Basic disk",
			disk: Disk{
				Label:      "boot",
				Size:       25000,
				Image:      "linode/ubuntu24.04",
				Filesystem: "ext4",
			},
			want: linodego.InstanceDiskCreateOptions{
				Label:      "boot",
				Size:       25000,
				Image:      "linode/ubuntu24.04",
				Filesystem: "ext4",
			},
		},
		{
			name: "Disk with authorized keys",
			disk: Disk{
				Label:          "boot",
				Size:           25000,
				Image:          "linode/arch",
				AuthorizedKeys: []string{"ssh-rsa AAAA..."},
			},
			want: linodego.InstanceDiskCreateOptions{
				Label:          "boot",
				Size:           25000,
				Image:          "linode/arch",
				AuthorizedKeys: []string{"ssh-rsa AAAA..."},
			},
		},
		{
			name: "Disk with stackscript",
			disk: Disk{
				Label:           "boot",
				Size:            25000,
				Image:           "linode/debian12",
				StackscriptID:   12345,
				StackscriptData: map[string]string{"key": "value"},
			},
			want: linodego.InstanceDiskCreateOptions{
				Label:           "boot",
				Size:            25000,
				Image:           "linode/debian12",
				StackscriptID:   12345,
				StackscriptData: map[string]string{"key": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenDisk(tt.disk)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("flattenDisk() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFlattenDisk_AllFields(t *testing.T) {
	disk := Disk{
		Label:           "boot",
		Size:            32000,
		Image:           "linode/ubuntu24.04",
		Filesystem:      "ext4",
		AuthorizedKeys:  []string{"ssh-rsa AAAA-test"},
		AuthorizedUsers: []string{"root", "devops"},
		StackscriptID:   123,
		StackscriptData: map[string]string{"foo": "bar"},
	}

	got := flattenDisk(disk)
	if got.Label != disk.Label {
		t.Fatalf("Label = %q, want %q", got.Label, disk.Label)
	}
	if got.Size != disk.Size {
		t.Fatalf("Size = %d, want %d", got.Size, disk.Size)
	}
	if got.Image != disk.Image {
		t.Fatalf("Image = %q, want %q", got.Image, disk.Image)
	}
	if got.Filesystem != disk.Filesystem {
		t.Fatalf("Filesystem = %q, want %q", got.Filesystem, disk.Filesystem)
	}
	if len(got.AuthorizedKeys) != 1 || got.AuthorizedKeys[0] != "ssh-rsa AAAA-test" {
		t.Fatalf("AuthorizedKeys = %v, want [ssh-rsa AAAA-test]", got.AuthorizedKeys)
	}
	if len(got.AuthorizedUsers) != 2 || got.AuthorizedUsers[0] != "root" || got.AuthorizedUsers[1] != "devops" {
		t.Fatalf("AuthorizedUsers = %v, want [root devops]", got.AuthorizedUsers)
	}
	if got.StackscriptID != 123 {
		t.Fatalf("StackscriptID = %d, want 123", got.StackscriptID)
	}
	if got.StackscriptData["foo"] != "bar" {
		t.Fatalf("StackscriptData = %v, want map[foo:bar]", got.StackscriptData)
	}
}

func TestFlattenInstanceConfigHelpers(t *testing.T) {
	t.Run("Nil helpers returns nil", func(t *testing.T) {
		result := flattenInstanceConfigHelpers(nil)
		if result != nil {
			t.Errorf("flattenInstanceConfigHelpers() expected nil for nil input")
		}
	})

	t.Run("All helpers set to true", func(t *testing.T) {
		trueVal := true
		helpers := &InstanceConfigHelpers{
			UpdateDBDisabled:  &trueVal,
			Distro:            &trueVal,
			ModulesDep:        &trueVal,
			Network:           &trueVal,
			DevTmpFsAutomount: &trueVal,
		}

		result := flattenInstanceConfigHelpers(helpers)
		if result == nil {
			t.Fatal("flattenInstanceConfigHelpers() unexpected nil result")
		}

		if !result.UpdateDBDisabled {
			t.Errorf("flattenInstanceConfigHelpers() UpdateDBDisabled = false, want true")
		}
		if !result.Distro {
			t.Errorf("flattenInstanceConfigHelpers() Distro = false, want true")
		}
		if !result.ModulesDep {
			t.Errorf("flattenInstanceConfigHelpers() ModulesDep = false, want true")
		}
		if !result.Network {
			t.Errorf("flattenInstanceConfigHelpers() Network = false, want true")
		}
		if !result.DevTmpFsAutomount {
			t.Errorf("flattenInstanceConfigHelpers() DevTmpFsAutomount = false, want true")
		}
	})

	t.Run("Mixed helper values", func(t *testing.T) {
		trueVal := true
		falseVal := false
		helpers := &InstanceConfigHelpers{
			UpdateDBDisabled:  &trueVal,
			Distro:            &falseVal,
			ModulesDep:        &trueVal,
			Network:           nil, // Not set
			DevTmpFsAutomount: &falseVal,
		}

		result := flattenInstanceConfigHelpers(helpers)
		if result == nil {
			t.Fatal("flattenInstanceConfigHelpers() unexpected nil result")
		}

		if !result.UpdateDBDisabled {
			t.Errorf("flattenInstanceConfigHelpers() UpdateDBDisabled = false, want true")
		}
		if result.Distro {
			t.Errorf("flattenInstanceConfigHelpers() Distro = true, want false")
		}
		if !result.ModulesDep {
			t.Errorf("flattenInstanceConfigHelpers() ModulesDep = false, want true")
		}
		if result.Network {
			t.Errorf("flattenInstanceConfigHelpers() Network = true, want false (nil defaults to false)")
		}
		if result.DevTmpFsAutomount {
			t.Errorf("flattenInstanceConfigHelpers() DevTmpFsAutomount = true, want false")
		}
	})
}

// TestFlattenInstanceConfig tests the integration of device resolution and config creation
func TestFlattenInstanceConfig(t *testing.T) {
	diskLabelToID := map[string]int{
		"boot": 12345,
		"swap": 67890,
	}

	tests := []struct {
		name        string
		config      InstanceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid config with devices",
			config: InstanceConfig{
				Label:    "my-config",
				Comments: "Test config",
				Devices: &InstanceConfigDevices{
					SDA: &InstanceConfigDevice{DiskLabel: "boot"},
					SDB: &InstanceConfigDevice{DiskLabel: "swap"},
				},
				Kernel:     "linode/grub2",
				RootDevice: "/dev/sda",
			},
			wantErr: false,
		},
		{
			name: "Invalid disk label in devices",
			config: InstanceConfig{
				Label: "my-config",
				Devices: &InstanceConfigDevices{
					SDA: &InstanceConfigDevice{DiskLabel: "nonexistent"},
				},
			},
			wantErr:     true,
			errContains: "failed to resolve devices",
		},
		{
			name: "Config with interfaces",
			config: InstanceConfig{
				Label: "my-config",
				Devices: &InstanceConfigDevices{
					SDA: &InstanceConfigDevice{DiskLabel: "boot"},
				},
				Interfaces: []Interface{
					{Purpose: "public"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := flattenInstanceConfig(tt.config, diskLabelToID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("flattenInstanceConfig() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("flattenInstanceConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("flattenInstanceConfig() unexpected error: %v", err)
				return
			}

			if opts.Label != tt.config.Label {
				t.Errorf("flattenInstanceConfig() Label = %v, want %v", opts.Label, tt.config.Label)
			}
			if opts.Comments != tt.config.Comments {
				t.Errorf("flattenInstanceConfig() Comments = %v, want %v", opts.Comments, tt.config.Comments)
			}

			if tt.config.Devices != nil {
				if tt.config.Devices.SDA != nil {
					if opts.Devices.SDA == nil || opts.Devices.SDA.DiskID != diskLabelToID[tt.config.Devices.SDA.DiskLabel] {
						t.Errorf("flattenInstanceConfig() SDA = %v, want disk id %d", opts.Devices.SDA, diskLabelToID[tt.config.Devices.SDA.DiskLabel])
					}
				}
				if tt.config.Devices.SDB != nil {
					if opts.Devices.SDB == nil || opts.Devices.SDB.DiskID != diskLabelToID[tt.config.Devices.SDB.DiskLabel] {
						t.Errorf("flattenInstanceConfig() SDB = %v, want disk id %d", opts.Devices.SDB, diskLabelToID[tt.config.Devices.SDB.DiskLabel])
					}
				}
			}

			if len(opts.Interfaces) != len(tt.config.Interfaces) {
				t.Errorf("flattenInstanceConfig() Interfaces length = %d, want %d", len(opts.Interfaces), len(tt.config.Interfaces))
			}
			for i := range tt.config.Interfaces {
				if i >= len(opts.Interfaces) {
					break
				}
				if opts.Interfaces[i].Purpose != linodego.ConfigInterfacePurpose(tt.config.Interfaces[i].Purpose) {
					t.Errorf("flattenInstanceConfig() Interfaces[%d].Purpose = %q, want %q", i, opts.Interfaces[i].Purpose, tt.config.Interfaces[i].Purpose)
				}
				if opts.Interfaces[i].Primary != tt.config.Interfaces[i].Primary {
					t.Errorf("flattenInstanceConfig() Interfaces[%d].Primary = %v, want %v", i, opts.Interfaces[i].Primary, tt.config.Interfaces[i].Primary)
				}
			}

			if tt.config.RootDevice == "" {
				if opts.RootDevice != nil {
					t.Errorf("flattenInstanceConfig() RootDevice = %v, want nil", opts.RootDevice)
				}
			} else {
				if opts.RootDevice == nil || *opts.RootDevice != tt.config.RootDevice {
					t.Errorf("flattenInstanceConfig() RootDevice = %v, want %q", opts.RootDevice, tt.config.RootDevice)
				}
			}

			if opts.MemoryLimit != tt.config.MemoryLimit || opts.Kernel != tt.config.Kernel || opts.InitRD != tt.config.InitRD || opts.RunLevel != tt.config.RunLevel || opts.VirtMode != tt.config.VirtMode {
				t.Errorf("flattenInstanceConfig() scalar fields mismatch: got %+v, config %+v", opts, tt.config)
			}
		})
	}
}

func TestFlattenInstanceConfig_AllFields(t *testing.T) {
	diskLabelToID := map[string]int{"boot": 101, "swap": 202}
	trueVal := true
	cfg := InstanceConfig{
		Label:    "cfg-all",
		Comments: "all fields",
		Devices: &InstanceConfigDevices{
			SDA: &InstanceConfigDevice{DiskLabel: "boot"},
			SDB: &InstanceConfigDevice{DiskLabel: "swap"},
		},
		Helpers: &InstanceConfigHelpers{
			UpdateDBDisabled:  &trueVal,
			Distro:            &trueVal,
			ModulesDep:        &trueVal,
			Network:           &trueVal,
			DevTmpFsAutomount: &trueVal,
		},
		Interfaces:  []Interface{{Purpose: "public", Primary: true}},
		MemoryLimit: 2048,
		Kernel:      "linode/grub2",
		InitRD:      44,
		RootDevice:  "/dev/sda",
		RunLevel:    "default",
		VirtMode:    "paravirt",
	}

	opts, err := flattenInstanceConfig(cfg, diskLabelToID)
	if err != nil {
		t.Fatalf("flattenInstanceConfig() unexpected error: %v", err)
	}

	if opts.Label != cfg.Label || opts.Comments != cfg.Comments {
		t.Fatalf("label/comments = %q/%q, want %q/%q", opts.Label, opts.Comments, cfg.Label, cfg.Comments)
	}
	if opts.Devices.SDA == nil || opts.Devices.SDA.DiskID != 101 {
		t.Fatalf("SDA = %v, want DiskID 101", opts.Devices.SDA)
	}
	if opts.Devices.SDB == nil || opts.Devices.SDB.DiskID != 202 {
		t.Fatalf("SDB = %v, want DiskID 202", opts.Devices.SDB)
	}
	if opts.Helpers == nil || !opts.Helpers.UpdateDBDisabled || !opts.Helpers.Distro || !opts.Helpers.ModulesDep || !opts.Helpers.Network || !opts.Helpers.DevTmpFsAutomount {
		t.Fatalf("helpers = %+v, want all true", opts.Helpers)
	}
	if len(opts.Interfaces) != 1 {
		t.Fatalf("interfaces length = %d, want 1", len(opts.Interfaces))
	}
	if opts.Interfaces[0].Purpose != linodego.ConfigInterfacePurpose("public") || !opts.Interfaces[0].Primary {
		t.Fatalf("interface mapping = %+v, want purpose public primary true", opts.Interfaces[0])
	}
	if opts.MemoryLimit != cfg.MemoryLimit || opts.Kernel != cfg.Kernel || opts.InitRD != cfg.InitRD || opts.RunLevel != cfg.RunLevel || opts.VirtMode != cfg.VirtMode {
		t.Fatalf("scalar config fields not mapped correctly: %+v", opts)
	}
	if opts.RootDevice == nil || *opts.RootDevice != cfg.RootDevice {
		t.Fatalf("RootDevice = %v, want %q", opts.RootDevice, cfg.RootDevice)
	}
}

// TestSelectBootConfig tests the boot configuration selection logic
func TestSelectBootConfig(t *testing.T) {
	tests := []struct {
		name              string
		configs           []InstanceConfig
		wantIndex         int
		wantError         bool
		wantErrorContains string
	}{
		{
			name: "Multiple configs with booted=true should error",
			configs: []InstanceConfig{
				{Label: "config1", Booted: true},
				{Label: "config2", Booted: true},
			},
			wantError:         true,
			wantErrorContains: "only one configuration profile can have 'booted' set to true",
		},
		{
			name: "Single config with booted=true should use that config",
			configs: []InstanceConfig{
				{Label: "config1", Booted: false},
				{Label: "config2", Booted: true},
				{Label: "config3", Booted: false},
			},
			wantIndex: 1,
			wantError: false,
		},
		{
			name: "No configs with booted=true should default to first config",
			configs: []InstanceConfig{
				{Label: "config1", Booted: false},
				{Label: "config2", Booted: false},
			},
			wantIndex: 0,
			wantError: false,
		},
		{
			name: "Single config not marked as booted should use that config",
			configs: []InstanceConfig{
				{Label: "config1", Booted: false},
			},
			wantIndex: 0,
			wantError: false,
		},
		{
			name: "First config marked as booted should return index 0",
			configs: []InstanceConfig{
				{Label: "config1", Booted: true},
				{Label: "config2", Booted: false},
			},
			wantIndex: 0,
			wantError: false,
		},
		{
			name: "Last config marked as booted should return its index",
			configs: []InstanceConfig{
				{Label: "config1", Booted: false},
				{Label: "config2", Booted: false},
				{Label: "config3", Booted: true},
			},
			wantIndex: 2,
			wantError: false,
		},
		{
			name: "Three configs with first one booted",
			configs: []InstanceConfig{
				{Label: "config1", Booted: true},
				{Label: "config2", Booted: false},
				{Label: "config3", Booted: false},
			},
			wantIndex: 0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, err := selectBootConfig(tt.configs)

			if tt.wantError {
				if err == nil {
					t.Errorf("selectBootConfig() expected error but got none")
					return
				}
				if tt.wantErrorContains != "" && !strings.Contains(err.Error(), tt.wantErrorContains) {
					t.Errorf("selectBootConfig() error = %v, want error containing %q", err, tt.wantErrorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("selectBootConfig() unexpected error: %v", err)
				return
			}

			if gotIndex != tt.wantIndex {
				t.Errorf("selectBootConfig() = %d, want %d (config: %q)", gotIndex, tt.wantIndex, tt.configs[tt.wantIndex].Label)
			}
		})
	}
}
