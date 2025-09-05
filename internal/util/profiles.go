package util

import (
        "path/filepath"
	"os"
)

// Profile struct is for handling lage data blobs
// ProfileID: an index to reference ProfileData, at the moment
//
//	this will be a sha256 caluculated from ProfileData
//	since, ProfileData can be any type (string, array
//	of strings, etc) it may need processing before the
//	sha256 can be calcuated
//
// ProfileData: can be any data such array of strings, a single large string
type Profile struct {
	ProfileID   string `json:"profileId"` // sha256 of ProfileData
	ProfileData any   `json:"profileData"`  // data associated with ProfileI
}

func GetChecksum(chkSumFilePath string) string {
	// before system registration chkSumFilePath won't exist
	// in that case just return emtpy string
	if !FileExists(chkSumFilePath) {
		return ""
	}
	// read value in from checksum file
	content, err := os.ReadFile(chkSumFilePath)
	// if read fails return empty string, should not happed
	if err != nil {
		return ""
	}
	return string(content)
}

func PutChecksum(chkSumFilePath string, value string) error {
        path := filepath.Dir(chkSumFilePath)
	perms := os.FileMode(0755)
		err := os.MkdirAll(path, perms)
	if err != nil {
		return err
	}
	perms = os.FileMode(0644)
	err = os.WriteFile(chkSumFilePath, []byte(value), perms)
	return err
}
