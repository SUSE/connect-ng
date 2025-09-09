package collectors

import (
	"bufio"
	"bytes"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/pkg/profiles"
)

type LSMOD struct {
	UpdateDataIDs bool
}

const kernModChecksumFile = "kernel-modules-profile-id"
const lsmodTag = "mod_list"

// Process output from lsmod to remove dups and get sorted list of kernel mods.
func getKernelMods(inputStream []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(inputStream))

	itemMap := make(map[string]bool)
	list := []string{}
	// Skip header line
	scanner.Scan()
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			item := strings.Fields(scanner.Text())[0]
			if _, exists := itemMap[item]; !exists {
				itemMap[item] = true
				list = append(list, item)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Strings(list)
	return list, nil
}

func (lsmod LSMOD) run(arch string) (Result, error) {
	util.Debug.Print("lsmod.UpdateDataIDs: ", lsmod.UpdateDataIDs)
	modInfo, err := util.Execute([]string{"lsmod"}, nil)
	if err != nil {
		return Result{}, err
	}

	sortedMods, _ := getKernelMods(modInfo)
	result, err := profiles.BuildProfile(lsmod.UpdateDataIDs, lsmodTag, kernModChecksumFile, sortedMods)

	return result, err
}
