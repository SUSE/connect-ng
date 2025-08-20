package connect

import (
	"encoding/json"
	"testing"
)

func TestDistroTarget(t *testing.T) {
	p := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
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

	p2 := Product{Name: "Dummy"}
	if !p2.isEmpty() {
		t.Errorf("expected %v to be empty", p2)
	}

	p3 := Product{Name: "sle-module-basesystem", Version: "15.2", Arch: "x86_64"}
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
	if p.IsBase {
		t.Error("Expected p.IsBase == false, got true")
	}
	if p.Extensions[0].Name != "extension1" {
		t.Errorf("Expected p.Extensions[0].Name == product1, got %s", p.Extensions[0].Name)
	}
	if p.Extensions[0].Available {
		t.Error("Expected p.Extensions[0].Available == false, got true")
	}
	if p.Extensions[0].IsBase {
		t.Error("Expected p.Extensions[0].IsBase == false, got true")
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
		t.Error("Expected p.Available == true, got false")
	}
	if p.Extensions[0].Name != "extension1" {
		t.Errorf("Expected p.Extensions[0].Name == product1, got %s", p.Extensions[0].Name)
	}
	if !p.Extensions[0].Available {
		t.Error("Expected p.Extensions[0].Available == true, got false")
	}
}

func TestUnmarshalJSONBase(t *testing.T) {
	jsn := `{"identifier": "product1", "base": true}`

	var p Product
	if err := json.Unmarshal([]byte(jsn), &p); err != nil {
		t.Errorf("Error unmarshalling: %s", err)
	}
	if !p.IsBase {
		t.Error("Expected p.IsBase == true, got false")
	}
}

func TestUnmarshalJSONIsBase(t *testing.T) {
	jsn := `{"identifier": "product1", "isbase": true}`

	var p Product
	if err := json.Unmarshal([]byte(jsn), &p); err != nil {
		t.Errorf("Error unmarshalling: %s", err)
	}
	if !p.IsBase {
		t.Error("Expected p.IsBase == true, got false")
	}
}

func TestUnmarshalJSONProductTypeBase(t *testing.T) {
	jsn := `{"identifier": "product1", "product_type": "base"}`

	var p Product
	if err := json.Unmarshal([]byte(jsn), &p); err != nil {
		t.Errorf("Error unmarshalling: %s", err)
	}
	if !p.IsBase {
		t.Error("Expected p.IsBase == true, got false")
	}
}

func TestSplitTriplet(t *testing.T) {
	expected := Product{Name: "a", Version: "b", Arch: "c"}
	p, err := SplitTriplet("a/b/c")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if p.ToTriplet() != expected.ToTriplet() {
		t.Errorf("Expected: %v, got: %v", expected, p)
	}
	_, err = SplitTriplet("SLES")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestFindIDInt(t *testing.T) {
	jsn := `{"id": 101361}`
	var p Product
	err := json.Unmarshal([]byte(jsn), &p)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if p.ID != 101361 {
		t.Errorf("Expected: %d, got: %d", 101361, p.ID)
	}
}

func TestFindIDString(t *testing.T) {
	jsn := `{"id": "101361"}`
	var p Product
	err := json.Unmarshal([]byte(jsn), &p)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if p.ID != 101361 {
		t.Errorf("Expected: %d, got: %d", 101361, p.ID)
	}
}

func TestMarshallProductIntID(t *testing.T) {
	p1 := Product{ID: 42}
	jsn, err := json.Marshal(&p1)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	p2 := Product{}
	err = json.Unmarshal([]byte(jsn), &p2)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if p1.ID != p2.ID {
		t.Errorf("Expected: %d, got: %d", p1.ID, p2.ID)
	}
}
