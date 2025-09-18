package collectors

import (
	"bufio"
	"bytes"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type LSMOD struct {
	UpdateDataIDs bool
}

const KernModChecksumFile = "kern-modules.txt"
const lsmodTag = "mod_list"

func (lsmod LSMOD) run(arch string) (Result, error) {
	util.Debug.Print("lsmod.UpdateDataIDs: ", lsmod.UpdateDataIDs)
	modInfo, err := util.Execute([]string{"lsmod"}, nil)
	// Should not happen, but if command fails
	// send nil to SCC to indicate that module data was unavailable.
	if err != nil {
		return Result{lsmodTag: nil}, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(modInfo))

	// Iterate through each line get module name
	modList := make(map[string]bool)
	// Skip header line
	scanner.Scan()
	for scanner.Scan() {
		module := strings.Fields(scanner.Text())[0]
		modList[module] = true // update module list
	}
	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		return Result{lsmodTag: nil}, err
	}

	// Run through module list removing duplicates
	var sortedMods []string
	for module := range modList {
		sortedMods = append(sortedMods, module)
	}

	// Sort the module list, , since the order does not matter
	// and may change over system reboots.
	sort.Strings(sortedMods)

	// Create new line seperated list of sorted modules
	output := strings.Join(sortedMods, "\n")

	// Get sha256 hash of pci data
	modProfileId := util.CalcSha256(output)

	oldmodProfileId := util.GetChecksum(KernModChecksumFile)
	if oldmodProfileId != modProfileId {
		if lsmod.UpdateDataIDs {
			util.Debug.Print("updating: ", KernModChecksumFile)
			err = util.PutChecksum(KernModChecksumFile, modProfileId)
			if err != nil {
				return Result{lsmodTag: nil}, err
			}
		}
		var dataBlob util.Profile
		dataBlob.ProfileID = modProfileId
		dataBlob.ProfileData = sortedMods
		return Result{lsmodTag: dataBlob}, nil
	} else {
		var dataBlob util.Profile
		dataBlob.ProfileID = modProfileId
		return Result{lsmodTag: dataBlob}, nil
	}
}
