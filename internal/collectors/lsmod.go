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

const kernModChecksumFile = "kern-modules.txt"
const lsmodTag = "mod_list"

// process output from lsmod to remove dups and get sorted list of kernel mods.
func getKernelMods(inputStream []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(inputStream))

	// Iterate through each line get module name
	itemMap := make(map[string]bool)
	// Skip header line
	scanner.Scan()
	for scanner.Scan() {
		item := strings.Fields(scanner.Text())[0]
		itemMap[item] = true // update item map
	}
	// Return nil and erroe if errors in scanning
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Run through module list removing duplicates
	var sortedList []string
	for module := range itemMap {
		sortedList = append(sortedList, module)
	}

	// Sort the list,
	sort.Strings(sortedList)
	return sortedList, nil
}

func (lsmod LSMOD) run(arch string) (Result, error) {
	util.Debug.Print("lsmod.UpdateDataIDs: ", lsmod.UpdateDataIDs)
	modInfo, err := util.Execute([]string{"lsmod"}, nil)
	// Should not happen, but if command fails
	// send nil Result if module data was unavailable.
	if err != nil {
		return Result{}, err
	}

	sortedMods, err := getKernelMods(modInfo)
	result, err := util.UpdateCache(lsmod.UpdateDataIDs, lsmodTag, kernModChecksumFile, sortedMods)

	return result, err
}
