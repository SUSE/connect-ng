package collectors

import (
	"bufio"
	"bytes"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

const pciChecksumFile = "pci-data-profile-id"
const pciTag = "pci_data"

type PCI struct {
	UpdateDataIDs bool
}

func getPciData(inputStream []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(inputStream))
	list := []string{}
	for scanner.Scan() {
		item := scanner.Text()
		list = append(list, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (pci PCI) run(arch string) (Result, error) {
	util.Debug.Print("pci.UpdateDataIDs: ", pci.UpdateDataIDs)
	output, err := util.Execute([]string{"lspci", "-s", ".0"}, nil)
	if err != nil {
		return Result{}, err
	}
	pciData, _ := getPciData(output)
	result, err := profiles.BuildProfile(pci.UpdateDataIDs, pciTag, pciChecksumFile, pciData)

	return result, err
}
