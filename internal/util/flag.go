package util

import (
	"strings"
)

type SliceFlag []string

func (sf *SliceFlag) String() string {
	return strings.Join(*sf, ",")
}

func (sf *SliceFlag) Set(value string) error {
	*sf = append(*sf, value)
	return nil
}
