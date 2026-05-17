package util

import (
	"bytes"
	"regexp"
)

//
// Systemd Unit Files
//

type SystemdUnitFile struct {
	Name string
}

func NewSystemdUnitFile(name string) *SystemdUnitFile {
	return &SystemdUnitFile{Name: name}
}

func (uf *SystemdUnitFile) String() string {
	return uf.Name
}

func (uf *SystemdUnitFile) Property(property string) ([]byte, error) {
	output, err := Systemctl(
		"show", "--property", property, uf.Name,
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

//
// List unit files that match specified type and state
//

var listUnitFilesOutputMatcher = regexp.MustCompile(
	`^(\S+)` + // match and extract first column, <unit_file_name>
		`\s+` + `(\S+)` + // match and extract second column, <state>
		`(?:` + // start a non-extracting group
		`\s+` + `\S+` + // match 3rd column, <preset> if present
		`)?$`, // optional as column not reported on older (SLE 12 SP5) systems
)

func ListMatchingUnitFilesOfTypeAndState(pattern, unitType, state string, nameFilter *regexp.Regexp) ([]*SystemdUnitFile, error) {
	output, err := Systemctl(
		"list-unit-files", "--type", unitType, "--state", state, pattern,
	)
	if err != nil {
		return nil, err
	}

	unitFiles := []*SystemdUnitFile{}
	for _, line := range bytes.Split(output, []byte("\n")) {
		// filter on lines that match target pattern
		enabledMatches := listUnitFilesOutputMatcher.FindSubmatch(line)

		// skip line if line doesn't match expected pattern, total line + 2 fields
		if len(enabledMatches) != 3 {
			continue
		}

		// skip line if state field is not desired state
		if string(enabledMatches[2]) != state {
			continue
		}

		// skip unit file name if it doesn't match the expected pattern
		if !nameFilter.Match(enabledMatches[1]) {
			continue
		}

		unitFiles = append(unitFiles, NewSystemdUnitFile(string(enabledMatches[1])))
	}

	return unitFiles, nil
}
