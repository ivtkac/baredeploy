package lsblk

import (
	"strings"
	"testing"
)

func TestColumns(t *testing.T) {
	cols := Columns()

	for _, want := range []string{"NAME", "SIZE", "PTTYPE", "MOUNTPOINT"} {
		if !strings.Contains(cols, want) {
			t.Errorf("Columns() = %q, missing %q", cols, want)
		}
	}
	if strings.Contains(cols, "children") {
		t.Errorf("Columns() should not contain nested-devices field: %q", cols)
	}
}

func TestParse(t *testing.T) {
	data := []byte(`{
		"blockdevices": [
			{
				"name": "sda", "type": "disk", "size": 512110190592,
				"model": "SAMSUNG MZ7LM512", "rota": "0", "mountpoint": null,
				"children": [
					{"name": "sda1", "type": "part", "size": 536870912, "mountpoint": "/boot"}
				]
			}
		]
	}`)

	out, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(out.BlockDevices) != 1 {
		t.Fatalf("got %d devices, want 1", len(out.BlockDevices))
	}

	dev := out.BlockDevices[0]
	if dev.Name != "sda" || dev.Type != "disk" {
		t.Errorf("unexpected device: %+v", dev)
	}
	if dev.Size == nil || *dev.Size != 512110190592 {
		t.Errorf("size not parsed: %v", dev.Size)
	}
	if dev.Rota == nil || bool(*dev.Rota) {
		t.Errorf("rota %q should parse as false", "0")
	}
	if len(dev.Children) != 1 || dev.Children[0].Name != "sda1" {
		t.Errorf("children not parsed: %+v", dev.Children)
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse([]byte("not json")); err == nil {
		t.Error("Parse() should fail on invalid JSON")
	}
}

func TestFlagUnmarshal(t *testing.T) {
	cases := map[string]bool{
		`"1"`: true, `"0"`: false, `true`: true, `false`: false,
	}
	for in, want := range cases {
		var f Flag
		if err := f.UnmarshalJSON([]byte(in)); err != nil {
			t.Fatalf("UnmarshalJSON(%s) error: %v", in, err)
		}
		if bool(f) != want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", in, f, want)
		}
	}
}
