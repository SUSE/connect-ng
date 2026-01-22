package profiles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/SUSE/connect-ng/internal/util"
)

type Result = map[string]interface{}

const clearCacheCount = "clear-cache-count"
const clearCacheCountLimit = 5

// WrappedProfile interface
type WrappedProfile interface {
	DeleteProfileCache(nameFilter string)
	PutCacheValue(name string, value string) error
	GetCacheValue(name string) string
}

// Where to store profile id cache files,  can changed via  SetProfileCache
// Can be changed with SetProfileFilePath
var profileFilePath = "/run/suseconnect/"

// ProfileCache struct definition
type ProfileCache struct {
}

// Global interface variable, can changed via  SetProfileCach.
var profileCache WrappedProfile = &ProfileCache{}

// Profile struct is for handling large data blobs
type Profile struct {
	Id   string `json:"identifier"`
	Data any    `json:"data"`
}

// ---------------------------------------------------------------------
// ProfileCache Methods (Implementing WrappedProfile Interface)
// ---------------------------------------------------------------------

// DeleteProfileCache clears data from profile cache that match
// specified filter
func (cache *ProfileCache) DeleteProfileCache(nameFilter string) {
	// based on filter remove files from profileFilePath ignoring errors
	filter := filepath.Join(GetProfileFilePath(), nameFilter)
	files, _ := filepath.Glob(filter)
	if len(files) != 0 {
		for _, file := range files {
			_ = os.Remove(file)
		}
	}
}

// PutCacheValue put name value pair into ProfileCache
func (cache *ProfileCache) PutCacheValue(name string, value string) error {
	path := filepath.Dir(GetProfileFilePath())
	perms := os.FileMode(0755)
	err := os.MkdirAll(path, perms)
	if err != nil {
		return err
	}

	filePath := filepath.Join(GetProfileFilePath(), name)
	perms = os.FileMode(0644)
	util.Debug.Println("PutCacheValue file, value: ", filePath, []byte(value))
	err = os.WriteFile(filePath, []byte(value), perms)
	return err
}

// GetCacheValue gets cached profile ID
func (cache *ProfileCache) GetCacheValue(name string) string {
	filePath := filepath.Join(GetProfileFilePath(), name)

	// Before system registration filePath won't exist
	if !util.FileExists(filePath) {
		return ""
	}
	// Read value in from checksum file
	content, err := os.ReadFile(filePath)
	// If read fails return empty string, should not happed
	if err != nil {
		return ""
	}
	return string(content)
}

func (cache *ProfileCache) Clear() {
	cache.DeleteProfileCache("*-profile-id")
	IncFailedProfileUpdate()
}

func (cache *ProfileCache) ResetClearCount() {
	ResetFailedProfileUpdate()
}

// ---------------------------------------------------------------------
// Utility Functions
// ---------------------------------------------------------------------

// Helper function to calculate SHA256 hash
func calcSha256(input string) string {
	mkSha256 := sha256.New()
	mkSha256.Write([]byte(input))
	sha256 := mkSha256.Sum(nil)
	return hex.EncodeToString(sha256)
}

// Helper function to determine if profile sending is disabled
func disableSendingProfile() bool {
	// The explicit type conversion is added for safety
	cnt, _ := strconv.Atoi(profileCache.GetCacheValue(clearCacheCount))
	return cnt > clearCacheCountLimit
}

// ---------------------------------------------------------------------
// Exported Functions
// ---------------------------------------------------------------------

// BuildProfile builds and returns a profile object
func BuildProfile(updateCache bool, tag string, cacheFile string, dataBlock any) (Result, error) {
	if disableSendingProfile() {
		return Result{}, nil
	}

	// Marshal to Json for ID calculation
	jsonBytes, err := json.Marshal(dataBlock)
	if err != nil {
		util.Debug.Printf("dataBlockType, : %T\n", dataBlock)
		util.Debug.Println("dataBlock, :", dataBlock)
		fmt.Printf("Failed: BuildProfile: %s, %v\n", "bad input:", err)
		return Result{}, err
	}
	hashInput := string(jsonBytes)
	cacheId := calcSha256(hashInput)

	var profile Profile
	profile.Id = cacheId

	savedCacheId := profileCache.GetCacheValue(cacheFile)
	if savedCacheId != cacheId {
		profile.Data = dataBlock
		if updateCache {
			util.Debug.Print("updating: ", cacheFile)
			err := profileCache.PutCacheValue(cacheFile, cacheId)
			if err != nil {
				fmt.Printf("Failed: Profile Cache Update: %s, %v\n", cacheFile, err)
				return Result{tag: profile}, err
			}
		}
	}
	// Use the tag as a key for the result map
	return Result{tag: profile}, nil
}

// DeleteProfileCache simply calls the interface method
func DeleteProfileCache(filter string) {
	util.Debug.Print("clearing ProfileCache with filter: ", filter)
	profileCache.DeleteProfileCache(filter)
}

// SetProfileFilePath uses the new SetPath method on the interface
func SetProfileFilePath(newPath string) {
	profileFilePath = filepath.Join(newPath, "/") + "/"
}

func GetProfileFilePath() string {
	return profileFilePath
}

func ResetFailedProfileUpdate() {
	util.Debug.Print("reset clearCacheCount: ")
	profileCache.PutCacheValue(clearCacheCount, "0")
}

func IncFailedProfileUpdate() {
	cnt, _ := strconv.Atoi(profileCache.GetCacheValue(clearCacheCount))
	cnt++
	util.Debug.Print("clearCacheCount set to: ", cnt)
	profileCache.PutCacheValue(clearCacheCount, strconv.Itoa(cnt))
}

// SetProfileCache is now correct as profileCache is of type WrappedProfile
func SetProfileCache(newCache WrappedProfile) {
	profileCache = newCache
}
