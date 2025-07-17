package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestPCIRun(t *testing.T) {
	assert := assert.New(t)

        expectedPCI := `
Slot:   0000:05:00.0
Class:  VGA compatible controller
Vendor: Advanced Micro Devices, Inc. [AMD/ATI]
Device: Rembrandt [Radeon 680M]
SVendor:        Advanced Micro Devices, Inc. [AMD/ATI]
SDevice:        Rembrandt [Radeon 680M]
Rev:    c7
ProgIf: 00
Driver: amdgpu
Module: amdgpu
IOMMUGroup:     17
`
        mockUtilExecute(expectedPCI, nil)
	expected := Result{"PCIData": expectedPCI}

	collector := PCI{}
	result, err := collector.run(ARCHITECTURE_X86_64)
        
	assert.Equal(expected, result)
	assert.Nil(err)
}
