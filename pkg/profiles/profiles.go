package profiles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SUSE/connect-ng/internal/util"
)

type Result = map[string]interface{}

// Where to store profile id cache files
// Can be changed with SetProfileFilePath
var profileFilePath = "/run/suseconnect/"

// Profile struct is for handling lage data blobs
// Id: an unique index to reference Profile data
//
//	Profile data can be any type (string, array
//
// Data: Can be any data such array of strings, a single large string,a map, etc.
type Profile struct {
	Id   string `json:"profileId"`
	Data any    `json:"profileData"`
}

//	BuildProfile builds and returns a profile object to be sent to the SCC or RMT servers
//
// Inputs:
//
//	updateCache: If true on, local profile cache ID will be updated
//	tag: tag to be sent to associated with data profile (i.e. "mod_list", pci_data",..)
//	cacheFile: name of the local profile cache ID file
//	dataBlock: data to proccessed
//
// Returns:
//
//	       Result record: Result = map[string]interface{}
//		  Error
//
// Given a data object of any type, the function contructs am unique ID for that data and builds
// a result record of the form: tag: "id":["id_value", "data": dataBlock].
// If the calculated ID maches local cached ID in the cacheFile, then the dataBlock is ommitted
// in the return.
// If updateCache is true and the alculated ID does not match the local cached ID in cacheFile,
// cacheFile is updated with the calculated ID value
func BuildProfile(updateCache bool, tag string, cacheFile string, dataBlock any) (Result, error) {
	// To build the ID we need a string eqivalent to dataBlock.
	// So, just mashal that to Json.
	json, err := json.Marshal(dataBlock)
	// Return nil Result if marshal failed
	if err != nil {
		util.Debug.Printf("dataBlockType, : %T\n", dataBlock)
		util.Debug.Println("dataBlock, :", dataBlock)
		fmt.Printf("Failed: BuildProfile: %s, %v\n", "bad input:", err)
		return Result{}, err
	}
	hashInput := string(json)

	// Use sha256 to build unique ID for dataBlock
	cacheId := calcSha256(hashInput)

	var profile Profile
	profile.Id = cacheId

	savedCacheId := GetCacheValue(cacheFile)
	if savedCacheId != cacheId {
		profile.Data = dataBlock
		if updateCache {
			util.Debug.Print("updating: ", cacheFile)
			err := PutCacheValue(cacheFile, cacheId)
			// Still return valid Result, because we can still send data even updating cache fails
			if err != nil {
				fmt.Printf("Failed: Profile Cache Update: %s, %v\n", cacheFile, err)
				return Result{tag: profile}, err
			}
		}
	}
	return Result{tag: profile}, nil
}

func GetCacheValue(chkSumFile string) string {
	// Only file nae passed in, need full file path
	chkSumFilePath := filepath.Join(profileFilePath, chkSumFile)

	// Before system registration chkSumFilePath won't exist
	// in that case just return emtpy string
	if !util.FileExists(chkSumFilePath) {
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

func calcSha256(input string) string {
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
