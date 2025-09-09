package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Result = map[string]interface{}

var profileFilePath = "/run/suseconnect/"

// Profile struct is for handling lage data blobs
// ProfileID: an index to reference ProfileData, at the moment
//
//	This will be a sha256 caluculated from ProfileData
//	since, ProfileData can be any type (string, array
//	of strings, etc) it may need processing before the
//	sha256 can be calcuated
//
// ProfileData: can be any data such array of strings, a single large string
type Profile struct {
	ProfileID   string `json:"profileId"`   // Sha256 of ProfileData
	ProfileData any    `json:"profileData"` // Data associated with ProfileI
}

func UpdateCache(updateCache bool, tag string, cacheFile string, dataBlock any) (Result, error) {

	var hashInput = ""
	switch v := dataBlock.(type) {
	case []string:
		hashInput = strings.Join(v, "\n")
	case string:
		hashInput = v
		// for all others, use string of encoded json to create has value
	default:
		Debug.Printf("unknown type for dataBlock, : %T\n", dataBlock)
		json, err := json.Marshal(dataBlock)
		// return nil Resultm if marshal failed
		if err != nil {
			return Result{}, err
		}
		hashInput = string(json)
	}
	// Get sha256 dataBlock

	cacheId := CalcSha256(hashInput)

	var profile Profile
	profile.ProfileID = cacheId

	savedCacheId := GetCacheValue(cacheFile)
	if savedCacheId != cacheId {
		profile.ProfileData = dataBlock
		if updateCache {
			Debug.Print("updating: ", cacheFile)
			err := PutCacheValue(cacheFile, cacheId)
			// return nill result if Put fails.
			if err != nil {
				return Result{}, err
			}
		}
	}
	return Result{tag: profile}, nil
}

func GetCacheValue(chkSumFile string) string {
	// Only file nae passed in, need full file path
	chkSumFilePath := profileFilePath + chkSumFile

	// Before system registration chkSumFilePath won't exist
	// in that case just return emtpy string
	if !FileExists(chkSumFilePath) {
		return ""
	}
	// Read value in from checksum file
	content, err := os.ReadFile(chkSumFilePath)
	// If read fails return empty string, should not happed
	if err != nil {
		return ""
	}
	return string(content)
}

func PutCacheValue(chkSumFile string, value string) error {
	path := filepath.Dir(profileFilePath)
	perms := os.FileMode(0755)
	err := os.MkdirAll(path, perms)
	if err != nil {
		return err
	}

	// Only file name passed in, need full file path
	chkSumFilePath := profileFilePath + chkSumFile
	perms = os.FileMode(0644)
	err = os.WriteFile(chkSumFilePath, []byte(value), perms)
	return err
}

func DeleteProfileCache() {
	// Remove profileFilePath and all files, ignore error
	_ = os.RemoveAll(profileFilePath)
}

func CalcSha256(input string) string {
	mkSha256 := sha256.New()
	mkSha256.Write([]byte(input))
	sha256 := mkSha256.Sum(nil)
	return hex.EncodeToString(sha256)
}

func SetProfileFilePath(newPath string) {
	profileFilePath = newPath
	if profileFilePath[len(profileFilePath)-1] != '/' {
		profileFilePath = profileFilePath + "/"
	}
}

func GetProfileFilePath() string {
	return profileFilePath
}
