package config

import (
	"errors"

	"github.com/knadh/koanf/maps"

	cli "github.com/urfave/cli/v2"
)

// Cliflag implements a urfave/cli command line provider.
type Cliflag struct {
	delim   string
	context *cli.Context
}

// Provider returns a commandline flags provider that returns
// a nested map[string]interface{} of environment variable where the
// nesting hierarchy of keys are defined by delim. For instance, the
// delim "." will convert the key `parent.child.key: 1`
// to `{parent: {child: {key: 1}}}`.
//
// It takes an optional (but recommended) Koanf instance to see if the
// the flags defined have been set from other providers, for instance,
// a config file. If they are not, then the default values of the flags
// are merged. If they do exist, the flag values are not merged but only
// the values that have been explicitly set in the command line are merged.
func Provider(c *cli.Context, delim string) *Cliflag {
	return &Cliflag{
		context: c,
		delim:   delim,
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Cliflag) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	for _, name := range p.context.FlagNames() {
		mp[name] = p.context.Generic(name)
	}
	return maps.Unflatten(mp, p.delim), nil
}

// ReadBytes is not supported by the env koanf.
func (p *Cliflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}

// Watch is not supported.
func (p *Cliflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("basicflag provider does not support this method")
}
