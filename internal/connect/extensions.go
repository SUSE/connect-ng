package connect

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/SUSE/connect-ng/internal/zypper"
	"github.com/SUSE/connect-ng/pkg/registration"
)

var (
	//go:embed extension.tmpl
	extensionTemplate string

	//go:embed extension-list.tmpl
	extensionListTemplate string

	// test method overwrites
	localIsRegistered = IsRegistered
	localBaseProduct  = zypper.BaseProduct
	localRootWritable = util.IsRootFSWritable
)

type extension struct {
	Name         string       `json:"identifier"`
	Version      string       `json:"version"`
	Arch         string       `json:"arch"`
	FriendlyName string       `json:"name"`
	Activated    bool         `json:"activated"`
	Available    bool         `json:"available"`
	Free         bool         `json:"free"`
	Extensions   []*extension `json:"extensions"`
}

func extensionTree(as []*registration.Activation, p *registration.Product) *extension {
	current := productToExtension(as, p)

	for _, extProduct := range p.Extensions {
		current.Extensions = append(current.Extensions, extensionTree(as, &extProduct))
	}
	return current
}

func productToExtension(as []*registration.Activation, p *registration.Product) *extension {
	return &extension{
		Name:         p.Name,
		Version:      p.Version,
		Arch:         p.Arch,
		FriendlyName: p.FriendlyName,
		Activated:    registration.ProductInActivations(p, as),
		Available:    p.Available,
		Free:         p.Free,
		Extensions:   []*extension{},
	}
}

func RenderExtensionTree(opts *Options, outputJson bool) (string, error) {
	// The system is registered remotely
	if !localIsRegistered() {
		return "", ErrListExtensionsUnregistered
	}

	base, err := localBaseProduct()
	if err != nil {
		return "", err
	}

	apiConnection := NewWrapper(opts)
	as, err := registration.FetchActivations(apiConnection.Connection)
	if err != nil {
		return "", err
	}

	product, err := registration.FetchProductInfo(apiConnection.Connection, base.Identifier, base.Version, base.Arch)
	if err != nil {
		return "", err
	}

	tree := extensionTree(as, product)

	if outputJson {
		result, err := json.Marshal(tree)
		return string(result), err
	}

	return renderText(tree, localRootWritable())
}

func indentBlock(spaces int, block string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(block, "\n", "\n"+pad, -1)
}

func renderText(tree *extension, writableRoot bool) (string, error) {
	tpl, _ := template.New("extensions-list").Parse(extensionListTemplate)
	output := bytes.Buffer{}

	// If the system is not writable we assume it is a transactional
	// system and show the appropriate command
	command := "suseconnect"
	if !writableRoot {
		command = "transactional-update register"
	}

	extensions, err := renderTextExtension(4, tree.Extensions, command)
	if err != nil {
		return "", err
	}

	if err := tpl.Execute(&output, extensions); err != nil {
		return "", err
	}
	return output.String(), nil
}

func renderTextExtension(indent int, exts []*extension, command string) ([]string, error) {
	tpl, _ := template.New("extension").Parse(extensionTemplate)
	all := []string{}

	for _, ext := range exts {
		output := bytes.Buffer{}
		code := fmt.Sprintf("%s/%s/%s", ext.Name, ext.Version, ext.Arch)

		args := struct {
			extension
			Command string
			Code    string
		}{extension: *ext, Command: command, Code: code}

		if err := tpl.Execute(&output, args); err != nil {
			return nil, err
		}
		all = append(all, indentBlock(indent, output.String()))
		leafs, err := renderTextExtension(indent+4, ext.Extensions, command)

		if err != nil {
			return nil, err
		}
		all = append(all, leafs...)
	}
	return all, nil
}
