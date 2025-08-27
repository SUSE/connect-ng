package collectors

import (
	"fmt"
	"testing"

	"github.com/SUSE/connect-ng/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestLSMODRunSuccess(t *testing.T) {
	assert := assert.New(t)

	var dataBlob util.Profile
	dataBlob.ProfileID = "a27b6fd9514de9b56319922c0af28783ceec4d36230e16a32d35b060fa733c33"
	dataBlob.ProfileData = []string{"acpi_tad", "autofs4", "ccp", "dmi_sysfs", "efi_pstore", "hid_generic", "i2c_algo_bit", "i2c_piix4", "i2c_smbus", "input_leds", "ip_tables", "joydev", "k10temp", "lp", "mac_hid", "mc", "msr", "nfnetlink", "nvme", "nvme_core", "parport", "parport_pc", "ppdev", "psmouse", "r8169", "rc_core", "realtek", "sch_fq_codel", "snd_pci_acp3x", "snd_soc_acpi", "soundcore", "usbhid", "x_tables"}

	kernModTestData := `rc_core                73728  1 cec
snd_soc_acpi           16384  3 snd_sof_amd_acp,snd_acp_config,snd_pci_ps
i2c_piix4              32768  0
mc                     81920  5 videodev,snd_usb_audio,videobuf2_v4l2,uvcvideo,videobuf2_common
k10temp                16384  0
i2c_algo_bit           16384  1 amdgpu
i2c_smbus              20480  1 i2c_piix4
soundcore              16384  1 snd
ccp                   155648  1 kvm_amd
snd_pci_acp3x          16384  0
input_leds             12288  0
joydev                 32768  0
acpi_tad               20480  0
mac_hid                12288  0
sch_fq_codel           24576  4
msr                    12288  0
parport_pc             53248  0
ppdev                  24576  0
lp                     28672  0
parport                73728  3 parport_pc,lp,ppdev
efi_pstore             12288  0
nfnetlink              20480  5 nft_compat,nf_tables,ip_set
dmi_sysfs              24576  0
ip_tables              32768  0
x_tables               65536  9 xt_conntrack,nft_compat,xt_tcpudp,xt_addrtype,xt_CHECKSUM,xt_set,ipt_REJECT,ip_tables,xt_MASQUERADE
autofs4                57344  2
hid_generic            12288  0
usbhid                 77824  0
nvme                   61440  2
psmouse               217088  0
r8169                 126976  0
nvme_core             225280  3 nvme
realtek                49152  2
`

	mockUtilExecute(kernModTestData, nil)
	expected := Result{"lsmod_data": dataBlob}

	collector := LSMOD{}
	result, err := collector.run(ARCHITECTURE_X86_64)

	assert.Equal(expected, result)
	assert.Nil(err)
}

func TestLSMODRunFail(t *testing.T) {
	assert := assert.New(t)

	mockUtilExecute("", fmt.Errorf("forced error"))
	expected := Result{"lsmod_data": nil}

	collector := LSMOD{}
	result, err := collector.run(ARCHITECTURE_X86_64)

	assert.Equal(expected, result)
	assert.ErrorContains(err, "forced error")
}
