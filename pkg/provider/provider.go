package provider

import (
	"github.com/spf13/pflag"
)

type Interface interface {
}

type NewFunc func(*pflag.FlagSet) (Interface, error)

type Provider struct {
	Name  string
	New   NewFunc
	Flags *pflag.FlagSet
}
