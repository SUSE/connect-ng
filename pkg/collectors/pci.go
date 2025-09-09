package collectors

import (
	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

const pciChecksumFile = "pci-data.txt"
const pciTag = "pci_data"

type PCI struct {
	UpdateDataIDs bool
}

func (pci PCI) run(arch string) (Result, error) {
	util.Debug.Print("pci.UpdateDataIDs: ", pci.UpdateDataIDs)
	output, err := util.Execute([]string{"lspci", "-s", ".0"}, nil)
	if err != nil {
		return Result{}, err
	}
	stringOutput := string(output)
	result, _ := profiles.BuildProfile(pci.UpdateDataIDs, pciTag, pciChecksumFile, stringOutput)

	return result, nil
}
