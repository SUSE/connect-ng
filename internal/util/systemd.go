package util

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SYSTEMD_BYPATTERNS_VERSION = 230
)

var (
	// error returned when a method is not available
	SystemdMethodNotAvailable = fmt.Errorf("method not available")
)

func availableMethodCheck(methodName string, tgtVersion, reqVersion uint64) error {
	// check if method is available from the target systemd
	if tgtVersion < reqVersion {
		return fmt.Errorf(
			"Systemd version %d required for method %s: %w",
			reqVersion, methodName, SystemdMethodNotAvailable,
		)
	}

	return nil
}

// parse the systemd version string to extract the initial "major"
// version value by striping off any trailing details leaving just
// the base integer version
func parseSystemdVersion(version string) (uint64, error) {
	major, _, _ := strings.Cut(version, ".")
	return strconv.ParseUint(major, 10, 64)
}

// SystemdUnitName represents a systemd unit name
type SystemdUnitName string

func (name SystemdUnitName) UnitType() string {
	ext := filepath.Ext(string(name))
	if ext == "" {
		return ""
	}
	return ext[1:]
}

func (name SystemdUnitName) UnitName() string {
	return filepath.Base(string(name))
}

func (name SystemdUnitName) Match(patterns ...string) bool {
	unitName := name.UnitName()
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, unitName); matched {
			return true
		}
	}

	return false
}

// SystemdUnitFile represents the systemd unit file as reported by Manager.ListUnitFiles* methods
type SystemdUnitFile struct {
	Name  SystemdUnitName
	State string
}

func NewSystemdUnitFile(name, state string) *SystemdUnitFile {
	unitFile := SystemdUnitFile{
		Name:  SystemdUnitName(name),
		State: state,
	}

	return &unitFile
}

// SystemdUnit represents the systemd unit as reported by Manager.ListUnits* methods
type SystemdUnit struct {
	Name        SystemdUnitName
	Description string
	LoadState   string
	ActiveState string
	SubState    string
	Following   string
	ObjectPath  string
	JobId       uint32
	JobType     string
	JobPath     string
}

func NewSystemdUnit(name SystemdUnitName, description, loadState, activeState, subState, following, objectPath string, jobId uint32, jobType, jobPath string) *SystemdUnit {
	unit := SystemdUnit{
		Name:        name,
		Description: description,
		LoadState:   loadState,
		ActiveState: activeState,
		SubState:    subState,
		Following:   following,
		ObjectPath:  objectPath,
		JobId:       jobId,
		JobType:     jobType,
		JobPath:     jobPath,
	}

	return &unit
}

// the expected number of fields per ExecStart entry
const SYSTEMD_EXEC_START_NUM_FIELDS = 10

// SystemdServiceExecStart represents the ExecStart property of a systemd service
type SystemdServiceExecStart struct {
	Path                    string
	Args                    []string
	IgnoreErrors            bool
	StartTimestamp          uint64
	StartTimestampMonotonic uint64
	ExitTimestamp           uint64
	ExitTimestampMonotonic  uint64
	Pid                     uint32
	ExitCode                int32
	ExitStatus              int32
}

func NewSystemdServiceExecStart(path string, args []string, ignoreErrors bool, startTimestamp, startTimestampMonotonic, exitTimestamp, exitTimestampMonotonic uint64, pid uint32, exitCode, exitStatus int32) *SystemdServiceExecStart {
	execStart := SystemdServiceExecStart{
		Path:                    path,
		Args:                    args,
		IgnoreErrors:            ignoreErrors,
		StartTimestamp:          startTimestamp,
		StartTimestampMonotonic: startTimestampMonotonic,
		ExitTimestamp:           exitTimestamp,
		ExitTimestampMonotonic:  exitTimestampMonotonic,
		Pid:                     pid,
		ExitCode:                exitCode,
		ExitStatus:              exitStatus,
	}

	return &execStart
}

// SystemdClient defines the interface we used for our systemd interactions
type SystemdClient interface {
	ListUnits() ([]*SystemdUnit, error)
	ListUnitsByPatterns(patterns ...string) ([]*SystemdUnit, error)
	ListUnitFiles() ([]*SystemdUnitFile, error)
	ListUnitFilesByPatterns(patterns ...string) ([]*SystemdUnitFile, error)
	GetExecStart(unitObjectPath string) ([]*SystemdServiceExecStart, error)
	GetSystemdVersion() (string, error)
	GetUnitFileState(unitName SystemdUnitName) (string, error)
	Close() error
}
