package collectors

import (
	"github.com/SUSE/connect-ng/internal/util"
)

type PCI struct{}

func (pci PCI) run(arch string) (Result, error) {
        output, err := util.Execute([]string{"lspci", "-vmmDk", "-s", ".0"}, nil)
	// Some systems may not have lspci command, so
	// if the execute fails, 
	// send nil to SCC to indicate that PCI data was unavailable.
	if err != nil {
		return Result{"PCIData": nil}, nil
	}
	
	return Result{"PCIData": string(output)}, nil
}
