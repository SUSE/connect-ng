package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FeatureTestEnv struct {
	REGCODE         string
	EXPIRED_REGCODE string
	HA_REGCODE      string

	CREDENTIALS_PATH string
}

func NewEnv(t *testing.T) *FeatureTestEnv {
	t.Helper()
	assert := assert.New(t)

	fetch := func(dest *string, key string) {
		value, found := os.LookupEnv(key)

		if !found {
			assert.FailNow(fmt.Sprintf("Can not fetch require environment variable `%s`. Make sure it is set correctly", key))
			return
		}
		*dest = value
	}

	env := &FeatureTestEnv{
		CREDENTIALS_PATH: "/etc/zypp/credentials.d",
	}

	// FIXME: Move to REGCODE in the CI. For now stay with
	//        VALID_REGCODE to not duplicate secrets too much
	fetch(&env.REGCODE, "REGCODE")
	fetch(&env.EXPIRED_REGCODE, "EXPIRED_REGCODE")
	fetch(&env.HA_REGCODE, "HA_REGCODE")

	return env
}

func (env *FeatureTestEnv) CredentialsPath(file string) string {
	return filepath.Join(env.CREDENTIALS_PATH, file)
}

func (env *FeatureTestEnv) allRegcodes() []string {
	return []string{
		env.REGCODE,
		env.EXPIRED_REGCODE,
		env.HA_REGCODE,
	}
}

func (env *FeatureTestEnv) RedactRegcodes(input string) string {
	for _, regcode := range env.allRegcodes() {
		input = strings.ReplaceAll(input, regcode, "[REDACTED]")
	}
	return input
}
