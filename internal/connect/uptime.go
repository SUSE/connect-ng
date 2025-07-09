package connect

import (
	"bufio"
	"errors"
	"os"
)

const (
	UptimeLogFilePath = "/etc/zypp/suse-uptime.log"
)

// readUptimeLogFile reads the system uptime log from a given file and
// returns them as a string array. If the given file does not exist,
// it will be interpreted as if the system uptime log feature is not
// enabled. Hence an empty array will be returned.
func readUptimeLogFile(uptimeLogFilePath string) ([]string, error) {
	// NOTE: the uptime log file is produced by the suse-uptime-tracker
	// (https://github.com/SUSE/uptime-tracker) service. If the service
	// is installed and enabled, barring any unforeseen errors, the
	// uptime log file is expected to there and updated on the regular
	// basis. If the service is not installed or otherwise disabled, the
	// uptime log file may not exist. In that case we assume the uptime
	// tracking feature is disabled.
	_, err := os.Stat(uptimeLogFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	uptimeLogFile, err := os.Open(uptimeLogFilePath)
	if err != nil {
		return nil, err
	}
	defer uptimeLogFile.Close()
	fileScanner := bufio.NewScanner(uptimeLogFile)
	var logEntries []string

	for fileScanner.Scan() {
		logEntries = append(logEntries, fileScanner.Text())
	}
	if err = fileScanner.Err(); err != nil {
		return nil, err
	}
	return logEntries, nil
}
