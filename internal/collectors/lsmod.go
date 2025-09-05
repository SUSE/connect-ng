package collectors

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type LSMOD struct {
	UpdateDataIDs bool
}

const KernModChecksumFile = "/run/suseconnect-chksum-kernmoddata.txt"
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
	for scanner.Scan() {
		module := strings.Fields(scanner.Text())[0]
		// skip "Mudule" from the hearder line
		if module != "Module" {
			modList[module] = true // update module list
		}
	}
	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		return Result{lsmodTag: nil}, err
	}

	// run through mudual list removing duplicates
	var sortedMods []string
	for module := range modList {
		sortedMods = append(sortedMods, module)
	}

	// sort the module list, , since the order does not matter
	// and may change over system reboots.
	sort.Strings(sortedMods)

	// create new line seperated list of sorted modules
	output := strings.Join(sortedMods, "\n")

	byteOutput := []byte(output)
	// get sha256 has a pci data
	mkSha256 := sha256.New()
	mkSha256.Write([]byte(byteOutput))
	sha256 := mkSha256.Sum(nil)
	modProfileId := hex.EncodeToString(sha256)

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
