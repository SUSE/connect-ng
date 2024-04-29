package collectors

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

type Memory struct{}

// Returns an integer with the amount of megabytes of total memory (i.e.
// `MemTotal` in /proc/meminfo). It will return 0 if this information could not
// be extracted for whatever reason.
func (Memory) run(arch Architecture) (Result, error) {
	fileContent, err := localOsReadfile("/proc/meminfo")
	if err != nil {
		util.Debug.Print("'/proc/meminfo' could not be read!")
		return NoResult, err
	}

	memInMBytes := parseMeminfo(bytes.NewReader(fileContent))
	return Result{"mem_total": memInMBytes}, nil
}

// Returns the parsed value for the given file. The implementation has been
// split from `systemMemory` so testing it is easier, but bear in mind that
// these two are coupled.
func parseMeminfo(file io.Reader) int {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || fields[0] != "MemTotal:" {
			continue
		}

		val, err := strconv.Atoi(fields[1])
		if err != nil {
			util.Debug.Printf("could not obtain memory information for this system: %v", err)
			return 0
		}
		return val / 1024
	}

	if scanner.Err() != nil {
		util.Debug.Printf("could not obtain memory information for this system: %v", scanner.Err())
		return 0
	}

	util.Debug.Print("could not obtain memory information for this system")
	return 0
}
