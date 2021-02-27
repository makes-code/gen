package utils

import (
	"strings"
)

type StringArray []string

func (sa StringArray) String() string {
	return strings.Join(sa, ", ")
}

func (sa *StringArray) Set(val string) error {
	*sa = append(*sa, val)
	return nil
}
