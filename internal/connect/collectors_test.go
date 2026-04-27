package connect

import (
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/internal/collectors"
	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/stretchr/testify/assert"
)

func TestFetchSystemInformation(t *testing.T) {
	tests := []struct {
		name          string
		arch          string
		mockArchFunc  func() (string, error)
		expectError   bool
		shouldHaveCPU bool
	}{
		{
			name:          "explicit architecture",
			arch:          collectors.ARCHITECTURE_X86_64,
			expectError:   false,
			shouldHaveCPU: true,
		},
		{
			name: "auto-detect architecture",
			arch: "",
			mockArchFunc: func() (string, error) {
				return collectors.ARCHITECTURE_X86_64, nil
			},
			expectError:   false,
			shouldHaveCPU: true,
		},
		{
			name: "architecture detection error",
			arch: "",
			mockArchFunc: func() (string, error) {
				return "", fmt.Errorf("architecture detection failed")
			},
			expectError:   true,
			shouldHaveCPU: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			SetCollectorConfig(collectors.NewCollectorOptions(map[string]collectorsconfig.CollectorConfig{}))

			if tt.mockArchFunc != nil {
				originalDetectArch := collectors.DetectArchitecture
				defer func() { collectors.DetectArchitecture = originalDetectArch }()
				collectors.DetectArchitecture = tt.mockArchFunc
			}

			result, err := FetchSystemInformation(tt.arch)

			if tt.expectError {
				assert.NotNil(err)
				assert.Equal(collectors.NoResult, result)
			} else {
				assert.Nil(err)
				assert.NotEqual(collectors.NoResult, result)

				if tt.shouldHaveCPU {
					_, hasCPUs := result["cpus"]
					assert.True(hasCPUs)
				}
			}
		})
	}
}

func TestFetchSystemInformationCollectorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]collectorsconfig.CollectorConfig
		checkFields map[string]bool // field name -> should exist
	}{
		{
			name: "mandatory collectors always run",
			config: map[string]collectorsconfig.CollectorConfig{
				"cpu": {Enabled: false},
			},
			checkFields: map[string]bool{
				"cpus": true, // CPU is mandatory, should exist despite config
			},
		},
		{
			name: "opt-in collector enabled",
			config: map[string]collectorsconfig.CollectorConfig{
				"sap": {Enabled: true},
			},
			checkFields: map[string]bool{
				"cpus": true, // Other collectors still run
			},
		},
		{
			name:   "only system info collectors run",
			config: map[string]collectorsconfig.CollectorConfig{},
			checkFields: map[string]bool{
				"cpus":     true,  // System info collector
				"pci_data": false, // Profile collector (pci_devices), should not run
				"mod_list": false, // Profile collector (kernel_modules), should not run
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			SetCollectorConfig(collectors.NewCollectorOptions(tt.config))

			result, err := FetchSystemInformation(collectors.ARCHITECTURE_X86_64)
			assert.Nil(err)

			for field, shouldExist := range tt.checkFields {
				_, exists := result[field]
				assert.Equal(shouldExist, exists, "field %s existence mismatch", field)
			}
		})
	}
}

func TestFetchSystemProfiles(t *testing.T) {
	tests := []struct {
		name         string
		arch         string
		updateCache  bool
		mockArchFunc func() (string, error)
		expectError  bool
	}{
		{
			name:        "explicit architecture with cache update",
			arch:        collectors.ARCHITECTURE_X86_64,
			updateCache: true,
			expectError: false,
		},
		{
			name:        "explicit architecture without cache update",
			arch:        collectors.ARCHITECTURE_X86_64,
			updateCache: false,
			expectError: false,
		},
		{
			name:        "auto-detect architecture",
			arch:        "",
			updateCache: true,
			mockArchFunc: func() (string, error) {
				return collectors.ARCHITECTURE_X86_64, nil
			},
			expectError: false,
		},
		{
			name:        "architecture detection error",
			arch:        "",
			updateCache: false,
			mockArchFunc: func() (string, error) {
				return "", fmt.Errorf("architecture detection failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			SetCollectorConfig(collectors.NewCollectorOptions(map[string]collectorsconfig.CollectorConfig{}))

			if tt.mockArchFunc != nil {
				originalDetectArch := collectors.DetectArchitecture
				defer func() { collectors.DetectArchitecture = originalDetectArch }()
				collectors.DetectArchitecture = tt.mockArchFunc
			}

			result, err := FetchSystemProfiles(tt.arch, tt.updateCache)

			if tt.expectError {
				assert.NotNil(err)
				assert.Equal(collectors.NoResult, result)
			} else {
				assert.Nil(err)
				assert.NotEqual(collectors.NoResult, result)
			}
		})
	}
}

func TestFetchSystemProfilesCollectorConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]collectorsconfig.CollectorConfig
		checkFields map[string]bool // field name -> should exist
	}{
		{
			name:   "default profile collectors run",
			config: map[string]collectorsconfig.CollectorConfig{},
			checkFields: map[string]bool{
				"pci_data": true,  // Profile collector (pci_devices), enabled by default
				"mod_list": true,  // Profile collector (kernel_modules), enabled by default
				"cpus":     false, // System info collector, should not run
			},
		},
		{
			name: "disable profile collector",
			config: map[string]collectorsconfig.CollectorConfig{
				"pci_devices": {Enabled: false},
			},
			checkFields: map[string]bool{
				"pci_data": false, // Disabled
				"mod_list": true,  // Still enabled
			},
		},
		{
			name: "disable all profile collectors",
			config: map[string]collectorsconfig.CollectorConfig{
				"pci_devices":    {Enabled: false},
				"kernel_modules": {Enabled: false},
			},
			checkFields: map[string]bool{
				"pci_data": false,
				"mod_list": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			SetCollectorConfig(collectors.NewCollectorOptions(tt.config))

			result, err := FetchSystemProfiles(collectors.ARCHITECTURE_X86_64, false)
			assert.Nil(err)

			for field, shouldExist := range tt.checkFields {
				_, exists := result[field]
				assert.Equal(shouldExist, exists, "field %s existence mismatch", field)
			}
		})
	}
}
