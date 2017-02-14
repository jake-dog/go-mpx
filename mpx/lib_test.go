package mpx

import "testing"

func TestStringParamDefault(t *testing.T) {
	var foo string
	result := StringParamDefault(foo, "bar")
	if result != "bar" {
		t.Error("Expected bar, got ", result)
	}
}
