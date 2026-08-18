// Package lsblk provides utilities for interacting with the `lsblk` command.
//
// See: https://man7.org/linux/man-pages/man8/lsblk.8.html
package lsblk

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// LsblkColumns defines every column that can be output by lsblk command.
//
// The lsblk tags on this struct are parsed into [Columns]
// function to construct the `-o` argument.
type LsblkColumns struct {
	// Device name
	Name string `lsblk:"NAME"`

	// Path to the device node
	Path string `lsblk:"PATH"`

	// Device type
	Type string `lsblk:"TYPE"`

	// Size of the device
	Size string `lsblk:"SIZE"`

	// Device identifier
	Model string `lsblk:"MODEL"`

	// Device vendor
	Vendor string `lsblk:"VENDOR"`

	// Disk serial number
	Serial string `lsblk:"SERIAL"`

	// Unique storage identifier
	WWN string `lsblk:"WWN"`

	// Device transport type
	Transport string `lsblk:"TRAN"`

	// Rotational device
	Rota string `lsblk:"ROTA"`

	// Removable device
	Removable string `lsblk:"RM"`

	// Read-only device
	ReadOnly string `lsblk:"RO"`

	// Physical sector size
	PhySec string `lsblk:"PHY-SEC"`

	// Logical sector size
	LogSec string `lsblk:"LOG-SEC"`

	// Filesystem type
	FsType string `lsblk:"FSTYPE"`

	// Partition table type
	PartType string `lsblk:"PTTYPE"`

	// Partition LABEL
	PartLabel string `lsblk:"PARTLABEL"`

	// Partition table identifier (usualy UUID)
	PartUUID string `lsblk:"PTUUID"`

	// Where the device is mounted
	MountPoint string `lsblk:"MOUNTPOINT"`
}

// Columns returns a comma-separated list of columns derived from the
// `lsblk` struct tags on [LsblkColumns].
//
// Any new fields added to [LsblkColumns] will automatically be included
// in the output and passed to `lsblk` as `-o` arguments.
func Columns() string {
	var cols LsblkColumns
	return columnList(cols)
}

// columnList extracts the `lsblk:"..."` tag values in [LsblkColumns]
func columnList(v any) string {
	t := derefType(v)
	var parts []string
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("lsblk"); tag != "" {
			parts = append(parts, tag)
		}
	}
	return strings.Join(parts, ",")
}

func derefType(v any) interface {
	NumField() int
	Field(int) reflect.StructField
} {
	return reflect.TypeOf(v)
}

// Output is the parsed output of `lsblk -J`.
type Output struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

// BlockDevice represents a block-device row from lsblk JSON.
type BlockDevice struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Type       string        `json:"type"`
	Size       *uint64       `json:"size"`
	Model      *string       `json:"model"`
	Vendor     *string       `json:"vendor"`
	Serial     *string       `json:"serial"`
	WWN        *string       `json:"wwn"`
	Transport  *string       `json:"tran"`
	Rota       *Flag         `json:"rota"`
	Removable  *Flag         `json:"rm"`
	ReadOnly   *Flag         `json:"ro"`
	PhySec     *int          `json:"phy-sec"`
	LogSec     *int          `json:"log-sec"`
	FsType     *string       `json:"fstype"`
	PartType   *string       `lsblk:"pttype"`
	PartLabel  *string       `lsblk:"partlabel"`
	PartUUID   *string       `lsblk:"partuuid"`
	MountPoint *string       `json:"mountpoint"`
	Children   []BlockDevice `json:"children"`
}

// Flag is a boolean that lsblk may emit as a JSON bool or the strings
// "0" or "1". Both forms are accepted udring unmarshalling.
type Flag bool

// Parse unmarshals raw `lsblk -J` stdout into typed structure [Output].
func Parse(data []byte) (*Output, error) {
	var out Output
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ToColumns converts a JSON-parsed [BlockDevice] into [LsblkColumns]
// representing all strings, suitable for display or further mapping.
func (dev *BlockDevice) ToColumns() LsblkColumns {
	return deviceToColumns(dev)
}

func deviceToColumns(dev *BlockDevice) LsblkColumns {
	return LsblkColumns{
		Name:       dev.Name,
		Path:       dev.Path,
		Type:       dev.Type,
		Size:       Uint64Str(dev.Size),
		Model:      strVal(dev.Model),
		Vendor:     strVal(dev.Vendor),
		Serial:     strVal(dev.Serial),
		WWN:        strVal(dev.WWN),
		Transport:  strVal(dev.Transport),
		Rota:       flagStr(dev.Rota),
		Removable:  flagStr(dev.Removable),
		ReadOnly:   flagStr(dev.ReadOnly),
		PhySec:     intStr(dev.PhySec),
		LogSec:     intStr(dev.LogSec),
		FsType:     strVal(dev.FsType),
		PartType:   strVal(dev.PartType),
		PartLabel:  strVal(dev.PartLabel),
		PartUUID:   strVal(dev.PartUUID),
		MountPoint: strVal(dev.MountPoint),
	}
}

func (f *Flag) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	*f = Flag(s == "1" || s == "true")
	return nil
}

func Uint64Str(i *uint64) string {
	if i == nil {
		return ""
	}
	return strconv.FormatUint(*i, 10)
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func intStr(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

func flagStr(f *Flag) string {
	if f == nil {
		return ""
	}
	if *f {
		return "1"
	}
	return "0"
}
