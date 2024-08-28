package main

import (
	"bufio"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	//go:embed version.txt
	version string
	//go:embed uptimeTrackerUsage.txt
	uptimeTrackerUsageText string
)

const (
	uptimeCheckLogsFilePath = "/etc/zypp/suse-uptime.log"
	dateStringFormat        = "2006-01-02"
	initUptimeHours         = "000000000000000000000000" // initialize the uptime hours bit string with
	daysBeforePurge         = 90                         // purge all the records after this many days
)

// getShortenedVersion returns the short program version
func getShortenedVersion() string {
	return strings.Split(strings.TrimSpace(version), "~")[0]
}

func exitOnError(err error) {
	if err == nil {
		return
	}
	fmt.Println(err)
	os.Exit(1)
}

func readUptimeLogFile(uptimeLogsFilePath string) (map[string]string, error) {
	uptimeLogsFile, err := os.Open(uptimeLogsFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file doesn't exist, so don't error out
			return nil, nil
		}
		return nil, err
	}
	fileScanner := bufio.NewScanner(uptimeLogsFile)
	fileScanner.Split(bufio.ScanLines)
	var logEntries = make(map[string]string)

	var entry []string
	for fileScanner.Scan() {
		entryText := fileScanner.Text()
		entry = strings.Split(entryText, ":")
		if len(entry) != 2 {
			return nil, errors.New("Uptime log file is corrupted. Invalid log entry " + entryText)
		}
		logEntries[entry[0]] = entry[1]
	}
	err = uptimeLogsFile.Close()
	if err != nil {
		return nil, err
	}
	return logEntries, nil
}

func purgeOldUptimeLog(uptimeLogs map[string]string) (map[string]string, error) {
	now := time.Now().UTC()
	purgeBefore := now.AddDate(0, 0, -daysBeforePurge)
	var purgedLogs = make(map[string]string)
	for day, uptimeHours := range uptimeLogs {
		timestamp, err := time.Parse(dateStringFormat, day)
		if err != nil {
			return nil, err
		}
		if timestamp.After(purgeBefore) {
			purgedLogs[day] = uptimeHours
		}
	}
	return purgedLogs, nil
}

func updateUptimeLog(uptimeLogs map[string]string) map[string]string {
	// NOTE: we are standardizing timezone to UTC
	now := time.Now().UTC()
	day := now.Format(dateStringFormat)
	hours, _, _ := now.Clock()
	_, ok := uptimeLogs[day]
	if !ok {
		uptimeLogs[day] = initUptimeHours
	}
	uptimeHoursMap := []rune(uptimeLogs[day])
	uptimeHoursMap[hours] = '1'
	uptimeLogs[day] = string(uptimeHoursMap)

	return uptimeLogs
}

func writeUptimeLogsFile(uptimeLogsFilePath string, uptimeLogs map[string]string) error {
	uptimeLogsFile, err := os.OpenFile(uptimeLogsFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer uptimeLogsFile.Close()

	// sort the keys
	keys := make([]string, 0, len(uptimeLogs))
	for day := range uptimeLogs {
		keys = append(keys, day)
	}
	sort.Strings(keys)

	for _, day := range keys {
		_, err = uptimeLogsFile.WriteString(day + ":" + uptimeLogs[day] + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var (
		version bool
	)

	flag.Usage = func() {
		fmt.Print(uptimeTrackerUsageText)
	}

	flag.BoolVar(&version, "version", false, "")

	flag.Parse()
	if version {
		fmt.Println(getShortenedVersion())
		os.Exit(0)
	}

	uptimeLogs, err := readUptimeLogFile(uptimeCheckLogsFilePath)
	exitOnError(err)
	uptimeLogs, err = purgeOldUptimeLog(uptimeLogs)
	exitOnError(err)
	uptimeLogs = updateUptimeLog(uptimeLogs)
	err = writeUptimeLogsFile(uptimeCheckLogsFilePath, uptimeLogs)
	exitOnError(err)
}
