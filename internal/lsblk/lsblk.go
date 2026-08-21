// Package lsblk provides utilities for interacting with the `lsblk` command.
//
// See: https://man7.org/linux/man-pages/man8/lsblk.8.html
package lsblk

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Output is the parsed output of `lsblk -J`.
type Output struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

// BlockDevice represents a block-device row from lsblk JSON.
//
// Each field carries two tags:
//   - `json` — the key emitted by `lsblk -J`
//   - `lsblk` — the column name passed to `lsblk -o` (see [Columns])
//
// Adding a new field with both tags is all that is needed to
// request and parse an extra column.
type BlockDevice struct {
	// Device name
	Name string `json:"name" lsblk:"NAME"`

	// Path to the device node
	Path string `json:"path" lsblk:"PATH"`

	// Device type (disk, part, rom, ...)
	Type string `json:"type" lsblk:"TYPE"`

	// Size of the device in bytes
	Size *uint64 `json:"size" lsblk:"SIZE"`

	// Device identifier
	Model *string `json:"model" lsblk:"MODEL"`

	// Device vendor
	Vendor *string `json:"vendor" lsblk:"VENDOR"`

	// Disk serial number
	Serial *string `json:"serial" lsblk:"SERIAL"`

	// Unique storage identifier
	WWN *string `json:"wwn" lsblk:"WWN"`

	// Device transport type
	Transport *string `json:"tran" lsblk:"TRAN"`

	// Rotational device
	Rota *Flag `json:"rota" lsblk:"ROTA"`

	// Removable device
	Removable *Flag `json:"rm" lsblk:"RM"`

	// Read-only device
	ReadOnly *Flag `json:"ro" lsblk:"RO"`

	// Physical sector size
	PhySec *int `json:"phy-sec" lsblk:"PHY-SEC"`

	// Logical sector size
	LogSec *int `json:"log-sec" lsblk:"LOG-SEC"`

	// Filesystem type
	FsType *string `json:"fstype" lsblk:"FSTYPE"`

	// Partition table type
	PartType *string `json:"pttype" lsblk:"PTTYPE"`

	// Partition label
	PartLabel *string `json:"partlabel" lsblk:"PARTLABEL"`

	// Partition table identifier (usually UUID)
	PartUUID *string `json:"ptuuid" lsblk:"PTUUID"`

	// Where the device is mounted
	MountPoint *string `json:"mountpoint" lsblk:"MOUNTPOINT"`

	// Nested devices (partitions, LVM volumes, ...)
	Children []BlockDevice `json:"children"`
}

// Columns returns a comma-separated list of column names derived from
// the `lsblk` struct tags on [BlockDevice], suitable for `lsblk -o`.
//
// Any new field added to [BlockDevice] with an `lsblk` tag is
// automatically included.
func Columns() string {
	t := reflect.TypeFor[BlockDevice]()
	parts := make([]string, 0, t.NumField())
	for field := range t.Fields() {
		if tag := field.Tag.Get("lsblk"); tag != "" {
			parts = append(parts, tag)
		}
	}
	return strings.Join(parts, ",")
}

// Parse unmarshals raw `lsblk -J` stdout into a typed [Output].
func Parse(data []byte) (*Output, error) {
	var out Output
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parsing lsblk output: %w", err)
	}
	return &out, nil
}

// Flag is a boolean that lsblk may emit as a JSON bool or the strings
// "0" or "1". Both forms are accepted during unmarshalling.
type Flag bool

func (f *Flag) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	*f = Flag(s == "1" || s == "true")
	return nil
}
