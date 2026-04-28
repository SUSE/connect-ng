package connect

import (
	"fmt"
	"os"
	"testing"

	"github.com/SUSE/connect-ng/internal/collectors"
	"github.com/SUSE/connect-ng/internal/util"
	collectorsconfig "github.com/SUSE/connect-ng/pkg/collectors"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Mock system commands for all tests
	util.Execute = func(cmd []string, _ []int) ([]byte, error) {
		if len(cmd) == 0 {
			return []byte(""), nil
		}

		switch cmd[0] {
		case "lspci":
			return []byte("00:00.0 Host bridge: Test Device\n"), nil
		case "lsmod":
			return []byte("Module                  Size  Used by\ntest_module            16384  0\n"), nil
		case "lscpu":
			return []byte("# CPU,Socket\n0,0\n1,0\n2,1\n3,1\n"), nil
		default:
			return []byte(""), nil
		}
	}

	os.Exit(m.Run())
}

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
		checkFields map[string]bool
	}{
		{
			name:   "mandatory collectors always run",
			config: map[string]collectorsconfig.CollectorConfig{"cpu": {Enabled: false}},
			checkFields: map[string]bool{
				"cpus": true,
			},
		},
		{
			name:   "only system info collectors run",
			config: map[string]collectorsconfig.CollectorConfig{},
			checkFields: map[string]bool{
				"cpus":     true,
				"pci_data": false,
				"mod_list": false,
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
		checkFields map[string]bool
	}{
		{
			name:   "default profile collectors run",
			config: map[string]collectorsconfig.CollectorConfig{},
			checkFields: map[string]bool{
				"pci_data": true,
				"mod_list": true,
				"cpus":     false,
			},
		},
		{
			name:   "disable profile collector",
			config: map[string]collectorsconfig.CollectorConfig{"pci_devices": {Enabled: false}},
			checkFields: map[string]bool{
				"pci_data": false,
				"mod_list": true,
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
