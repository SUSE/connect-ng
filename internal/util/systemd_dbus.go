package util

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	DBUS_SYSTEMD_DEST    = "org.freedesktop.systemd1"
	DBUS_SYSTEMD_PATH    = "/org/freedesktop/systemd1"
	DBUS_SYSTEMD_MANAGER = DBUS_SYSTEMD_DEST + ".Manager"
	DBUS_SYSTEMD_SERVICE = DBUS_SYSTEMD_DEST + ".Service"
)

// DbusSystemdClient implements a SystemdClient that leverages native
// D-Bus support to talk to the system bus.
type DbusSystemdClient struct {
	conn        *dbus.Conn
	version     uint64
	fullVersion string
}

func (c *DbusSystemdClient) Init(conn *dbus.Conn) error {
	var err error

	// setup the connection so that GetSustemVersion() can be called
	c.conn = conn

	// retrieve the Systemd Version string which may have one of these forms
	//   * 228 (SLE 12 SP 5)
	//   * 246.16+suse.312.gb540e1826d (SLE 15 SP3)
	//   * 249.17+suse.247.g8b6ed60a0c (SLE 15 SP4/5)
	//   * 254.27+suse.192.gc89ea566d9 (SLE 15 SP6/7)
	//   * 257.13+suse.38.gd349fc5cd4 (SLE 16.0)
	c.fullVersion, err = c.GetSystemdVersion()
	if err != nil {
		return err
	}

	if c.version, err = parseSystemdVersion(c.fullVersion); err != nil {
		return err
	}

	return nil
}

// NewDbusSystemdClient connects to the system bus
func NewDbusSystemdClient() (*DbusSystemdClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	systemdClient := new(DbusSystemdClient)
	err = systemdClient.Init(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return systemdClient, nil
}

func (c *DbusSystemdClient) Close() error {
	return c.conn.Close()
}

//
// Implement ListUnits* methods
//

// type signature for unit info returned by ListUnits* methods: ssssssouso
type unitInfo struct {
	Name, Desc, Load, Active, Sub, Follow string
	ObjectPath                            dbus.ObjectPath
	JobID                                 uint32
	JobType                               string
	JobPath                               dbus.ObjectPath
}

func (c *DbusSystemdClient) ListUnits() ([]*SystemdUnit, error) {
	method := DBUS_SYSTEMD_MANAGER + ".ListUnits"
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, DBUS_SYSTEMD_PATH)

	var resp []unitInfo
	err := systemd.Call(method, 0).Store(&resp)
	if err != nil {
		return nil, err
	}

	var units []*SystemdUnit
	for _, u := range resp {
		units = append(units, NewSystemdUnit(
			SystemdUnitName(u.Name), u.Desc, u.Load, u.Active, u.Sub, u.Follow,
			string(u.ObjectPath), u.JobID, u.JobType, string(u.JobPath),
		))
	}
	return units, nil
}

func (c *DbusSystemdClient) ListUnitsByPatterns(patterns ...string) ([]*SystemdUnit, error) {
	methodName := "ListUnitsByPatterns"

	// check if method is available
	if err := availableMethodCheck(methodName, c.version, SYSTEMD_BYPATTERNS_VERSION); err != nil {
		return nil, err
	}

	method := DBUS_SYSTEMD_MANAGER + "." + methodName
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, DBUS_SYSTEMD_PATH)

	var resp []unitInfo
	err := systemd.Call(method, 0, []string{}, patterns).Store(&resp)
	if err != nil {
		return nil, err
	}

	var units []*SystemdUnit
	for _, u := range resp {
		units = append(units, NewSystemdUnit(
			SystemdUnitName(u.Name), u.Desc, u.Load, u.Active, u.Sub, u.Follow,
			string(u.ObjectPath), u.JobID, u.JobType, string(u.JobPath),
		))
	}
	return units, nil
}

//
// Implement ListUnitFiles* methods
//

// type signature for unit info returned by ListUnitFiles* methods: ss
type unitFileInfo struct {
	Name  string
	State string
}

func (c *DbusSystemdClient) ListUnitFiles() ([]*SystemdUnitFile, error) {
	method := DBUS_SYSTEMD_MANAGER + ".ListUnitFiles"
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, DBUS_SYSTEMD_PATH)

	var resp []unitFileInfo
	err := systemd.Call(method, 0).Store(&resp)
	if err != nil {
		return nil, err
	}

	var units []*SystemdUnitFile
	for _, u := range resp {
		units = append(units, NewSystemdUnitFile(u.Name, u.State))
	}
	return units, nil
}

func (c *DbusSystemdClient) ListUnitFilesByPatterns(patterns ...string) ([]*SystemdUnitFile, error) {
	methodName := "ListUnitFilesByPatterns"

	// check if method is available
	if err := availableMethodCheck(methodName, c.version, SYSTEMD_BYPATTERNS_VERSION); err != nil {
		return nil, err
	}

	method := DBUS_SYSTEMD_MANAGER + "." + methodName
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, DBUS_SYSTEMD_PATH)

	// type signature for unit info returned by ListUnitFilesByPattern: ss
	type unitFileInfo struct {
		Name  string
		State string
	}

	var resp []unitFileInfo
	err := systemd.Call(method, 0, []string{}, patterns).Store(&resp)
	if err != nil {
		return nil, err
	}

	var units []*SystemdUnitFile
	for _, u := range resp {
		units = append(units, NewSystemdUnitFile(u.Name, u.State))
	}
	return units, nil
}

func (c *DbusSystemdClient) GetUnitFileState(unitName SystemdUnitName) (string, error) {
	method := DBUS_SYSTEMD_MANAGER + ".GetUnitFileState"
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, DBUS_SYSTEMD_PATH)

	var resp string
	err := systemd.Call(method, 0, unitName).Store(&resp)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (c *DbusSystemdClient) GetExecStart(unitObjectPath string) ([]*SystemdServiceExecStart, error) {
	property := DBUS_SYSTEMD_SERVICE + ".ExecStart"
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, dbus.ObjectPath(unitObjectPath))

	// signature of the ExecStart property value is: a(sasbttttuii)
	variant, err := systemd.GetProperty(property)
	if err != nil {
		return nil, err
	}

	execStartEntries, ok := variant.Value().([][]any)
	if !ok {
		return nil, fmt.Errorf("unexpected ExecStart signature")
	}

	var results []*SystemdServiceExecStart
	for _, entry := range execStartEntries {
		// there should be at least 10 fields per entry
		if len(entry) < SYSTEMD_EXEC_START_NUM_FIELDS {
			continue
		}

		// convert the 10 fields per the signature: sasbttttuii
		path, ok0 := entry[0].(string)
		args, ok1 := entry[1].([]string)
		ignoreErrors, ok2 := entry[2].(bool)
		startTimestamp, ok3 := entry[3].(uint64)
		startTimestampMonotonic, ok4 := entry[4].(uint64)
		exitTimestamp, ok5 := entry[5].(uint64)
		exitTimestampMonotonic, ok6 := entry[6].(uint64)
		pid, ok7 := entry[7].(uint32)
		exitCode, ok8 := entry[8].(int32)
		exitStatus, ok9 := entry[9].(int32)

		// skip if any field failed to convert
		if !ok0 || !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 || !ok7 || !ok8 || !ok9 {
			continue
		}

		results = append(results, NewSystemdServiceExecStart(
			path, args, ignoreErrors,
			startTimestamp, startTimestampMonotonic,
			exitTimestamp, exitTimestampMonotonic,
			pid, exitCode, exitStatus,
		))
	}

	return results, nil
}

func (c *DbusSystemdClient) GetSystemdVersion() (string, error) {
	property := DBUS_SYSTEMD_MANAGER + ".Version"
	systemd := c.conn.Object(DBUS_SYSTEMD_DEST, dbus.ObjectPath(DBUS_SYSTEMD_PATH))

	// signature of the Version property value is: s
	variant, err := systemd.GetProperty(property)
	if err != nil {
		return "", err
	}

	systemdVersion, ok := variant.Value().(string)
	if !ok {
		return "", fmt.Errorf("unexpected Version signature")
	}

	return systemdVersion, nil
}

// verify that DbusSystemdClient implements the SystemdClient interface
var _ SystemdClient = (*DbusSystemdClient)(nil)
