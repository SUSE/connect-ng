package collectors

import (
	"fmt"
	"os"
	"testing"

	"github.com/SUSE/connect-ng/pkg/profiles"
	"github.com/stretchr/testify/assert"
)

var pciDataBlob profiles.Profile
var pciTestData string

func setupPciTestData() {
	testProfilePath, _ := os.MkdirTemp("/tmp/", "__suseconnect")
	profiles.SetProfileFilePath(testProfilePath + "/")

	pciDataBlob.Id = "8912df9878ef02c80c33dd530cb0005768fcc7606d37def01e27f1f5b20ba1da"

	pciDataBlob.Data = []string{
		"00:00.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Root Complex (rev 01)",
		"00:01.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)",
		"00:02.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)",
		"00:03.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)",
		"00:04.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)",
		"00:08.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)",
		"00:14.0 SMBus: Advanced Micro Devices, Inc. [AMD] FCH SMBus Controller (rev 71)",
		"00:18.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Rembrandt Data Fabric: Device 18h; Function 0",
		"01:00.0 Non-Volatile memory controller: Micron/Crucial Technology Device 5426 (rev 01)",
		"02:00.0 Network controller: Intel Corporation Wi-Fi 6 AX200 (rev 1a)",
		"03:00.0 Ethernet controller: Realtek Semiconductor Co., Ltd. RTL8111/8168/8211/8411 PCI Express Gigabit Ethernet Controller (rev 15)",
		"04:00.0 Ethernet controller: Realtek Semiconductor Co., Ltd. RTL8111/8168/8211/8411 PCI Express Gigabit Ethernet Controller (rev 15)",
		"05:00.0 VGA compatible controller: Advanced Micro Devices, Inc. [AMD/ATI] Rembrandt [Radeon 680M] (rev c7)",
		"06:00.0 USB controller: Advanced Micro Devices, Inc. [AMD] Rembrandt USB4 XHCI controller #8"}

	pciTestData = "00:00.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Root Complex (rev 01)\n00:01.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)\n00:02.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)\n00:03.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)\n00:04.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)\n00:08.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Family 17h-19h PCIe Dummy Host Bridge (rev 01)\n00:14.0 SMBus: Advanced Micro Devices, Inc. [AMD] FCH SMBus Controller (rev 71)\n00:18.0 Host bridge: Advanced Micro Devices, Inc. [AMD] Rembrandt Data Fabric: Device 18h; Function 0\n01:00.0 Non-Volatile memory controller: Micron/Crucial Technology Device 5426 (rev 01)\n02:00.0 Network controller: Intel Corporation Wi-Fi 6 AX200 (rev 1a)\n03:00.0 Ethernet controller: Realtek Semiconductor Co., Ltd. RTL8111/8168/8211/8411 PCI Express Gigabit Ethernet Controller (rev 15)\n04:00.0 Ethernet controller: Realtek Semiconductor Co., Ltd. RTL8111/8168/8211/8411 PCI Express Gigabit Ethernet Controller (rev 15)\n05:00.0 VGA compatible controller: Advanced Micro Devices, Inc. [AMD/ATI] Rembrandt [Radeon 680M] (rev c7)\n06:00.0 USB controller: Advanced Micro Devices, Inc. [AMD] Rembrandt USB4 XHCI controller #8"
}

func TestPCIRunSuccessNoUpdate(t *testing.T) {
	assert := assert.New(t)
	setupPciTestData()
	mockUtilExecute(pciTestData, nil)
	expected := Result{pciTag: pciDataBlob}

	collector := PCI{UpdateDataIDs: false}
	result, err := collector.run(ARCHITECTURE_X86_64)
	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestPCIRunSuccessUpdate(t *testing.T) {
	assert := assert.New(t)

	mockUtilExecute(pciTestData, nil)
	expected := Result{pciTag: pciDataBlob}

	collector := PCI{UpdateDataIDs: true}
	result, err := collector.run(ARCHITECTURE_X86_64)

	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestPCIRunSumsMatch(t *testing.T) {
	assert := assert.New(t)

	mockUtilExecute(pciTestData, nil)

	collector := PCI{UpdateDataIDs: true}
	result, err := collector.run(ARCHITECTURE_X86_64)
	fmt.Println(result)

	var expectedDataBlob profiles.Profile
	expectedDataBlob.Id = pciDataBlob.Id
	expected := Result{pciTag: expectedDataBlob}

	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestPCIRunFail(t *testing.T) {
	assert := assert.New(t)

	mockUtilExecute("", fmt.Errorf("forced error"))
	expected := Result{}

	collector := PCI{}
	result, err := collector.run(ARCHITECTURE_X86_64)

	profiles.DeleteProfileCache("*")
	assert.Equal(expected, result)
	assert.ErrorContains(err, "forced error")
}
