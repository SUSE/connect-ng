package util

const (
	MOCK_SYSTEMD_SLE12SP5_VERSION = "228"
	MOCK_SYSTEMD_SLE15SP3_VERSION = "246.16+suse.312.gb540e1826d"
)

// MockSystemdClient allows injecting fake data for testing.
type MockSystemdClient struct {
	// method mocking
	CloseFunc                   func() error
	ListUnitsFunc               func() ([]*SystemdUnit, error)
	ListUnitsByPatternsFunc     func(patterns ...string) ([]*SystemdUnit, error)
	ListUnitFilesFunc           func() ([]*SystemdUnitFile, error)
	ListUnitFilesByPatternsFunc func(patterns ...string) ([]*SystemdUnitFile, error)
	GetExecStartFunc            func(unitObjectPath string) ([]*SystemdServiceExecStart, error)
	GetSystemdVersionFunc       func() (string, error)
	GetUnitFileStateFunc        func(unitName SystemdUnitName) (string, error)

	// local variables
	version     uint64
	fullVersion string
}

func NewMockSystemdClient(fullVersion string) *MockSystemdClient {
	client := new(MockSystemdClient)

	if fullVersion == "" {
		fullVersion = MOCK_SYSTEMD_SLE15SP3_VERSION
	}
	client.fullVersion = fullVersion
	client.version, _ = parseSystemdVersion(fullVersion)

	return client
}

func (m *MockSystemdClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockSystemdClient) ListUnits() ([]*SystemdUnit, error) {
	if m.ListUnitsFunc != nil {
		return m.ListUnitsFunc()
	}
	return nil, nil
}

func (m *MockSystemdClient) ListUnitsByPatterns(patterns ...string) ([]*SystemdUnit, error) {
	// check if method is available
	if err := availableMethodCheck("ListUnitsByPatterns", m.version, SYSTEMD_BYPATTERNS_VERSION); err != nil {
		return nil, err
	}

	if m.ListUnitsByPatternsFunc != nil {
		return m.ListUnitsByPatternsFunc(patterns...)
	}
	return nil, nil
}

func (m *MockSystemdClient) ListUnitFiles() ([]*SystemdUnitFile, error) {
	if m.ListUnitFilesFunc != nil {
		return m.ListUnitFilesFunc()
	}
	return nil, nil
}

func (m *MockSystemdClient) ListUnitFilesByPatterns(patterns ...string) ([]*SystemdUnitFile, error) {
	// check if method is available
	if err := availableMethodCheck("ListUnitFilesByPatterns", m.version, SYSTEMD_BYPATTERNS_VERSION); err != nil {
		return nil, err
	}

	if m.ListUnitFilesByPatternsFunc != nil {
		return m.ListUnitFilesByPatternsFunc(patterns...)
	}
	return nil, nil
}

func (m *MockSystemdClient) GetExecStart(unitObjectPath string) ([]*SystemdServiceExecStart, error) {
	if m.GetExecStartFunc != nil {
		return m.GetExecStartFunc(unitObjectPath)
	}
	return nil, nil
}

func (m *MockSystemdClient) GetSystemdVersion() (string, error) {
	if m.GetSystemdVersionFunc != nil {
		return m.GetSystemdVersionFunc()
	}
	return "", nil
}

func (m *MockSystemdClient) GetUnitFileState(unitName SystemdUnitName) (string, error) {
	if m.GetUnitFileStateFunc != nil {
		return m.GetUnitFileStateFunc(unitName)
	}
	return "", nil
}

// verify that MockSystemdClient implements the SystemdClient interface
var _ SystemdClient = (*MockSystemdClient)(nil)
