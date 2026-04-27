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
				"cpus":           true,  // System info collector
				"pci_devices":    false, // Profile collector, should not run
				"kernel_modules": false, // Profile collector, should not run
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
