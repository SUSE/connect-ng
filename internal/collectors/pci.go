package collectors

import (
	"github.com/SUSE/connect-ng/internal/util"
)

const pciChecksumFile = "pci-devices.txt"
const pciTag = "pci_data"

type PCI struct {
	UpdateDataIDs bool
}

func (pci PCI) run(arch string) (Result, error) {
	util.Debug.Print("pci.UpdateDataIDs: ", pci.UpdateDataIDs)
	output, err := util.Execute([]string{"lspci", "-s", ".0"}, nil)
	// Some systems may not have lspci command, so
	// if the execute fails,
	// returb nil Resukt if PCI data was unavailable.
	if err != nil {
		return Result{}, err
	}
	stringOutput := string(output)
	result, err := util.UpdateCache(pci.UpdateDataIDs, pciTag, pciChecksumFile, stringOutput)

	return result, err
}
