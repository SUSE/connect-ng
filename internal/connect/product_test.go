package connect

import (
	"encoding/json"
	"testing"
)

func TestDistroTarget(t *testing.T) {
	p := NewProduct("sle-module-basesystem", "15.2", "x86_64")
	got := p.distroTarget()
	expect := "sle-15-x86_64"
	if got != expect {
		t.Errorf("got: %s, expected: %s", got, expect)
	}
}

func TestEmptyProduct(t *testing.T) {
	p1 := Product{}
	if !p1.isEmpty() {
		t.Errorf("expected %v to be empty", p1)
	}

	p2 := NewProduct("Dummy", "", "")
	if !p2.isEmpty() {
		t.Errorf("expected %v to be empty", p2)
	}

	p3 := NewProduct("sle-module-basesystem", "15.2", "x86_64")
	if p3.isEmpty() {
		t.Errorf("expected %v not to be empty", p3)
	}
}

func TestUnmarshalJSONsmt(t *testing.T) {
	jsn := `{"identifier": "product1", "available": true,
               "extensions": [{"identifier": "extension1", "available": false}]}`
	var p Product
	if err := json.Unmarshal([]byte(jsn), &p); err != nil {
		t.Errorf("Error unmarshalling: %s", err)
	}
	if p.Name != "product1" {
		t.Errorf("Expected p.Name == product1, got %s", p.Name)
	}
	if !p.Available {
		t.Error("Expected p.Aailable == true, got false")
	}
	if p.Extensions[0].Name != "extension1" {
		t.Errorf("Expected p.Extensions[0].Name == product1, got %s", p.Extensions[0].Name)
	}
	if p.Extensions[0].Available {
		t.Error("Expected p.Extensions[0].Available == false, got true")
	}
}

func TestUnmarshalJSONscc(t *testing.T) {
	jsn := `{"identifier": "product1",  "extensions": [{"identifier": "extension1"}]}`

	var p Product
	if err := json.Unmarshal([]byte(jsn), &p); err != nil {
		t.Errorf("Error unmarshalling: %s", err)
	}
	if p.Name != "product1" {
		t.Errorf("Expected p.Name == product1, got %s", p.Name)
	}
	if !p.Available {
		t.Error("Expected p.Aailable == true, got false")
	}
	if p.Extensions[0].Name != "extension1" {
		t.Errorf("Expected p.Extensions[0].Name == product1, got %s", p.Extensions[0].Name)
	}
	if !p.Extensions[0].Available {
		t.Error("Expected p.Extensions[0].Available == true, got false")
	}
}
