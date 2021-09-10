package connect

import (
	"testing"
)

func TestStringSet(t *testing.T) {
	ss := NewStringSet()
	ss.Add("cat")
	ss.Add("dog")
	ss.Add("dog")
	if ss.Len() != 2 {
		t.Fatal("Len() should be 2")
	}
	if !ss.Contains("dog") {
		t.Fatal("Should have found dog")
	}

	ss2 := NewStringSet("dog", "cat", "mouse")
	if ss2.Len() != 3 {
		t.Fatal("Len() should be 3")
	}
	ss2.Add("bird", "fish")
	if ss2.Len() != 5 {
		t.Fatal("Len() should be 5")
	}
	ss2.Delete("cat")
	if ss2.Contains("cat") {
		t.Fatal("Should not have found cat")
	}
	if ss2.Len() != 4 {
		t.Fatal("Len() should be 4")
	}
}
