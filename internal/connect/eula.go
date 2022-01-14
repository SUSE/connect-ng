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
)

const eulaIndex = "directory.yast"

var (
	eulaMatch = regexp.MustCompile(`^license\.(.*)\.txt$`)
)

func selectEULALanguage(index map[string]string) string {
	if len(index) == 0 {
		return ""
	}
	lang := CFG.Language
	// remove the encoding (e.g. ".UTF-8")
	lang = strings.Split(lang, ".")[0]
	// default to en_US
	if lang == "" || lang == "C" || lang == "POSIX" {
		lang = "en_US"
	}
	// language priorities (xx_YY, xx, en_US, en, ...)
	prios := []string{lang, strings.Split(lang, "_")[0], "en_US", "en"}
	for _, l := range prios {
		if _, ok := index[l]; ok {
			return l
		}
	}
	// last option: return first available language (sorted to make things consistent)
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
		Debug.Printf("Ignoring unknown index entry: %s", name)
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
	Debug.Printf("Downloading license index from %s...", url)
	index, err := downloadFile(url.String())
	if err != nil {
		return map[string]string{}, err
	}
	ret, err := parseEULAIndex(index, baseURL)
	if err != nil {
		return map[string]string{}, err
	}
	Debug.Printf("Downloaded license index %+v", ret)
	return ret, nil
}

func showTextInPager(text []byte) error {
	// POSIX default
	pager := "more"
	if p, ok := os.LookupEnv("PAGER"); ok && p != "" {
		pager = p
	}
	Debug.Printf("Using %s as pager.", pager)

	// run interactive pager (`sh` is needed to allow custom pagers e.g. `less -N`)
	comm := exec.Command("sh", "-c", pager)
	comm.Stdin = bytes.NewReader(text)
	comm.Stdout = os.Stdout
	comm.Stderr = os.Stderr
	return comm.Run()
}

func printEULA(text []byte) {
	// there is some padding at the end which doesn't make much sense outside of pager
	Info.Println(string(bytes.TrimSpace(text)))
}

// AcceptEULA displays EULA and prompts for acceptance (or does nothing if there's no EULA)
func AcceptEULA(autoAgree bool) error {
	// separate EULAs are only for addon products
	if CFG.Product.isEmpty() {
		return nil
	}

	// fetch list of extensions and search for requested product
	base, err := baseProduct()
	if err != nil {
		return err
	}
	prod, err := showProduct(base)
	if err != nil {
		return err
	}
	extension, _ := prod.findExtension(CFG.Product)
	// no EULA or product not found (handle in registration code)
	if strings.TrimSpace(extension.EULAURL) == "" {
		return nil
	}

	eulas, err := downloadEULAIndex(extension.EULAURL)
	if err != nil {
		return err
	}
	lang := selectEULALanguage(eulas)
	// this should happen only if there are no license files in index
	if lang == "" {
		return fmt.Errorf("No EULAs found at: %s", extension.EULAURL)
	}

	// download EULA text
	eula, err := downloadFile(eulas[lang])
	if err != nil {
		return err
	}
	// trim BOM (some pagers display it as hex codes)
	eula = bytes.TrimLeft(eula, "\xef\xbb\xbf")
	// add header
	header := fmt.Sprintf(
		"In order to install '%s', you must agree to terms of the following license agreement:\n\n",
		extension.FriendlyName)
	eula = append([]byte(header), eula...)
	if autoAgree {
		printEULA(eula)
		// TODO: maybe print some message e.g. "License accepted (auto confirmation)"
		return nil
	}
	// TODO: maybe also add navigation and exit hints like in
	// https://github.com/openSUSE/zypper/blob/master/src/utils/pager.cc
	if err := showTextInPager(eula); err != nil {
		// print EULA on pager failure
		printEULA(eula)
	}
	// prompt for acceptance
	var answer string
	for answer != "yes" && answer != "no" {
		Info.Print("Do you agree with the terms of the license? (yes/[no]):")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("\nStandard input seems to be closed, please restart " +
				"the operation in interactive mode or use '--auto-agree-with-licenses' option.")
		}
		answer = strings.ToLower(strings.TrimSpace(scanner.Text()))
		// default if ENTER pressed
		if answer == "" {
			answer = "no"
		}
	}
	if answer != "yes" {
		return fmt.Errorf("aborting installation due to user disagreement with %s license",
			extension.FriendlyName)
	}
	return nil
}
