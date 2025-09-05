package collectors

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/SUSE/connect-ng/internal/util"
)

// var InfoOnly bool // indicated that -i/--info was used
const PCIChecksumFile = "/run/suseconnect-chksum-pcidata.txt"
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
	// get sha256 has a pci data
	mkSha256 := sha256.New()
	mkSha256.Write([]byte(output))
	sha256 := mkSha256.Sum(nil)
	hwIndex := hex.EncodeToString(sha256)

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
