package linode

import (
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

			if got.Label != tt.want.Label {
				t.Errorf("flattenDisk() Label = %v, want %v", got.Label, tt.want.Label)
			}
			if got.Size != tt.want.Size {
				t.Errorf("flattenDisk() Size = %v, want %v", got.Size, tt.want.Size)
			}
			if got.Image != tt.want.Image {
				t.Errorf("flattenDisk() Image = %v, want %v", got.Image, tt.want.Image)
			}
			if got.Filesystem != tt.want.Filesystem {
				t.Errorf("flattenDisk() Filesystem = %v, want %v", got.Filesystem, tt.want.Filesystem)
			}
		})
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
		})
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
