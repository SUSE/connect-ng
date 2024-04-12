package credentials

import "strings"

// handleSystemToken stores the given token into the system credentials file
// unHless it's blank.
func HandleSystemToken(token string, fsRoot string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	creds, err := ReadCredentials(SystemCredentialsPath(fsRoot))
	if err != nil {
		return err
	}

	creds.SystemToken = token
	return creds.write()
}
