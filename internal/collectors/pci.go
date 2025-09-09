package collectors

import (
	"github.com/SUSE/connect-ng/internal/util"
)

const PCIChecksumFile = "pci-devices.txt"
const pciTag = "pci_data"

type PCI struct {
	UpdateDataIDs bool
}

func (pci PCI) run(arch string) (Result, error) {
	util.Debug.Print("pci.UpdateDataIDs: ", pci.UpdateDataIDs)
	output, err := util.Execute([]string{"lspci", "-s", ".0"}, nil)
	// Some systems may not have lspci command, so
	// if the execute fails,
	// send nil to SCC to indicate that PCI data was unavailable.
	if err != nil {
		return Result{pciTag: nil}, err
	}
	stringOutput := string(output)
	// Get sha256 has a pci data
	hwIndex := util.CalcSha256(string(output))

	oldHwIndex := util.GetChecksum(PCIChecksumFile)
	if oldHwIndex != hwIndex {
		if pci.UpdateDataIDs {
			util.Debug.Print("updating: ", PCIChecksumFile)
			err = util.PutChecksum(PCIChecksumFile, hwIndex)
			if err != nil {
				return Result{pciTag: nil}, err
			}
		}
		var dataBlob util.Profile
		dataBlob.ProfileID = hwIndex
		dataBlob.ProfileData = stringOutput
		return Result{pciTag: dataBlob}, nil
	} else {
		var dataBlob util.Profile
		dataBlob.ProfileID = hwIndex
		return Result{pciTag: dataBlob}, nil
	}
}
