package connect

import (
	"bytes"
	_ "embed" // golint
	"sort"
	"strings"
	"text/template"
)

var (
	//go:embed extensions-list.tmpl
	extensionsListTemplate string
)

// helper struct to simplify extensions list template
type displayExtension struct {
	Product   Product
	Code      string
	Activated bool

	Indent     string
	ConnectCmd string
}

// GetExtensionsList returns the text output for --list-extensions
func GetExtensionsList() (string, error) {
	if !IsRegistered() {
		return "", ErrListExtensionsUnregistered
	}
	base, err := baseProduct()
	if err != nil {
		return "", err
	}
	activations, err := systemActivations()
	if err != nil {
		return "", err
	}
	if _, ok := activations[base.ToTriplet()]; !ok {
		return "", ErrListExtensionsUnregistered
	}
	product, err := showProduct(base)
	if err != nil {
		return "", err
	}
	return printExtensions(product.Extensions, activations, isRootFSWritable())
}

func printExtensions(extensions []Product, activations map[string]Activation, rootFSWritable bool) (string, error) {
	t, err := template.New("extensions-list").Parse(extensionsListTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	cmd := "SUSEConnect"
	if !rootFSWritable {
		cmd = "transactional-update register"
	}
	err = t.Execute(&output, preformatExtensions(extensions, activations, cmd, 1))
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// this function takes a tree of products and returns a flattened version
// with some additional info to make the output template as simple as possible
func preformatExtensions(extensions []Product, activations map[string]Activation, cmd string, level int) []displayExtension {
	// sort (copy of) input by name
	sorted := make([]Product, len(extensions))
	copy(sorted, extensions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FriendlyName < sorted[j].FriendlyName
	})

	var ret []displayExtension
	for _, e := range sorted {
		_, activated := activations[e.ToTriplet()]
		ret = append(ret, displayExtension{
			Product:    e,
			Code:       e.ToTriplet(),
			Activated:  activated,
			Indent:     strings.Repeat("    ", level),
			ConnectCmd: cmd,
		})
		// add subextensions
		ret = append(ret, preformatExtensions(e.Extensions, activations, cmd, level+1)...)
	}
	return ret
}
