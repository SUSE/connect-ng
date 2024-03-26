package connect

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/SUSE/connect-ng/internal/util"
)

const eulaIndex = "directory.yast"

var (
	eulaMatch = regexp.MustCompile(`^license\.(.*)\.txt$`)
)

func selectEULALanguage(index map[string]string) string {
	if len(index) == 0 {
		return ""
	}

	// Remove the encoding (e.g. ".UTF-8") and default to en_US.
	lang := CFG.Language
	lang = strings.Split(lang, ".")[0]
	if lang == "" || lang == "C" || lang == "POSIX" {
		lang = "en_US"
	}

	// Language priorities (xx_YY, xx, en_US, en, ...)
	prios := []string{lang, strings.Split(lang, "_")[0], "en_US", "en"}
	for _, l := range prios {
		if _, ok := index[l]; ok {
			return l
		}
	}

	// Last option: return first available language (sorted to make things consistent)
	langs := make([]string, 0)
	for l := range index {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	return langs[0]
}

func parseEULAIndex(data []byte, baseURL string) (map[string]string, error) {
	ret := make(map[string]string, 0)
	base, err := url.Parse(baseURL)
	if err != nil {
		return ret, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		name := scanner.Text()
		if name == eulaIndex {
			continue
		}
		url, _ := base.Parse(name)
		if name == "license.txt" {
			ret["en_US"] = url.String()
			continue
		}
		match := eulaMatch.FindStringSubmatch(name)
		if len(match) == 2 {
			ret[match[1]] = url.String()
			continue
		}
		util.Debug.Printf("Ignoring unknown index entry: %s", name)
	}
	return ret, nil
}

func downloadEULAIndex(baseURL string) (map[string]string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return map[string]string{}, err
	}
	url, err := base.Parse(eulaIndex)
	if err != nil {
		return map[string]string{}, err
	}
	util.Debug.Printf("Downloading license index from %s...", url)
	index, err := downloadFile(url.String())
	if err != nil {
		return map[string]string{}, err
	}
	ret, err := parseEULAIndex(index, baseURL)
	if err != nil {
		return map[string]string{}, err
	}
	util.Debug.Printf("Downloaded license index %+v", ret)
	return ret, nil
}

func showTextInPager(text []byte) error {
	// POSIX default
	pager := "more"
	if p, ok := os.LookupEnv("PAGER"); ok && p != "" {
		pager = p
	}
	util.Debug.Printf("Using %s as pager.", pager)

	// Run interactive pager (`sh` is needed to allow custom pagers e.g. `less
	// -N`)
	comm := exec.Command("sh", "-c", pager)
	comm.Stdin = bytes.NewReader(text)
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stderr
	return comm.Run()
}

// Prints the text of the EULA without some padding at the end which doesn't
// make much sense outside of the pager.
func printEULA(text []byte) {
	util.Info.Println(string(bytes.TrimSpace(text)))
}

// Returns the text of the EULA to be shown to the user, or nothing if there was
// an error.
func fetchEULAFrom(extension Product) ([]byte, error) {
	eulas, err := downloadEULAIndex(extension.EULAURL)
	if err != nil {
		return nil, err
	}
	lang := selectEULALanguage(eulas)

	// This should happen only if there are no license files in index
	if lang == "" {
		return nil, fmt.Errorf("No EULAs found at: %s", extension.EULAURL)
	}

	// Download EULA text
	eula, err := downloadFile(eulas[lang])
	if err != nil {
		return nil, err
	}

	// Trim BOM (some pagers display it as hex codes)
	eula = bytes.TrimLeft(eula, "\xef\xbb\xbf")

	// Add header
	header := fmt.Sprintf(
		"In order to install '%s', you must agree to terms of the following license agreement:\n\n",
		extension.FriendlyName)

	return append([]byte(header), eula...), nil
}

// Prompts the user to accept or not the EULA. If there was an error or the user
// did not accept the EULA an error will be returned, otherwise this function
// returns nil.
func promptUser(text []byte, extensionName string) error {
	// Show the text itself into the pager. If the pager is not available, then
	// just print it.
	if err := showTextInPager(text); err != nil {
		printEULA(text)
	}

	for {
		// Using `fmt` instead of `Info` to avoid an unexpected (and ugly)
		// newline after printing this line.
		fmt.Print("Do you agree with the terms of the license? [y/yes n/no] (n): ")

		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("\nStandard input seems to be closed, please restart " +
				"the operation in interactive mode or use '--auto-agree-with-licenses' option.")
		}

		answer = strings.ToLower(strings.TrimSpace(answer))

		if answer == "y" || answer == "yes" {
			return nil
		} else if answer == "" || answer == "n" || answer == "no" {
			return fmt.Errorf("aborting installation due to user disagreement with %s license", extensionName)
		}
	}
}

// AcceptEULA handles EULA interactions from the given base product/extensions.
// If there are no EULAs to be handled, then this function does nothing and
// returns nil. If the global configuration is set to auto-accept EULAs, then
// they are accepted without further action. Otherwise, we prompt to users to
// confirm it and then we proceed accordingly.
func AcceptEULA() error {
	// Separate EULAs are only for addon products
	if CFG.Product.isEmpty() {
		return nil
	}

	// Fetch list of extensions and search for requested product
	base, err := baseProduct()
	if err != nil {
		return nil
	}

	// Skip if we can not fetch the product information and thus the EULA since there
	// might be no registration in place right now.
	// See bsc#1218649 and bsc#1217961
	prod, err := showProduct(base)
	if err != nil {
		util.Debug.Print("Cannot load base product details for EULA.")
		return nil
	}
	extension, _ := prod.findExtension(CFG.Product)

	// No EULA or product not found (handle in registration code)
	if strings.TrimSpace(extension.EULAURL) == "" {
		util.Debug.Printf("No EULA found for '%s'.", CFG.Product.Name)
		return nil
	}

	// Fetch the text from the EULA if possible.
	eula, err := fetchEULAFrom(extension)
	if err != nil {
		return err
	}

	if CFG.AutoAgreeEULA {
		printEULA(eula)
		fmt.Println("")
		util.Info.Printf("-> License auto-accepted through user configuration ...\n\n")
		return nil
	}

	return promptUser(eula, extension.FriendlyName)
}
