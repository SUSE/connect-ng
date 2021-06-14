package connect

import (
	"testing"
)

var testProducts = `<?xml version='1.0'?>
<stream>
  <product-list>
    <product name="SUSE-MicroOS" version="5.0" release="1" epoch="0" arch="x86_64" vendor="SUSE" summary="SUSE Linux Enterprise Micro 5.0" repo="@System" productline="SUSE-MicroOS" registerrelease="" shortname="SUSE Linux Enterprise Micro" flavor="" isbase="true" installed="true"><endoflife time_t="1606694400" text="2020-11-30T00:00:00Z"/><registerflavor/><description>SUSE Linux Enterprise Micro 5.0</description></product>
    <product name="suse-openstack-cloud" version="8" release="0" epoch="0" arch="x86_64" vendor="SUSE LINUX GmbH, Nuernberg, Germany" summary="SUSE OpenStack Cloud 8" repo="@System" productline="suse-openstack-cloud" registerrelease="" shortname="SOC8" flavor="POOL" isbase="false" installed="true"><endoflife time_t="1622419200" text="2021-05-31T00:00:00+00"/><registerflavor>extension</registerflavor><description>SUSE OpenStack Cloud 8</description></product>
  </product-list>
</stream>
`

func TestParseProductsXML(t *testing.T) {
	products, err := parseProductsXML([]byte(testProducts))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].toTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].toTriplet())
	}
}
