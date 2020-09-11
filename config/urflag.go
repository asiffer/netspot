package config

import (
	"errors"

	"github.com/knadh/koanf/maps"

	cli "github.com/urfave/cli/v2"
)

// Cliflag implements a urfave/cli command line provider.
type Cliflag struct {
	mp map[string]interface{}
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
	mp := make(map[string]interface{})
	for _, name := range c.FlagNames() {
		mp[name] = c.Value(name)
		// need to convert to raw go types
		// for some urfave/cli objects
		switch mp[name].(type) {
		case cli.StringSlice:
			ss, _ := mp[name].(cli.StringSlice)
			mp[name] = ss.Value()
		default:
			// pass
		}
	}
	return &Cliflag{
		mp: maps.Unflatten(mp, delim),
	}
}

// Read reads the flag variables and returns a nested conf map.
func (p *Cliflag) Read() (map[string]interface{}, error) {
	return p.mp, nil
}

// ReadBytes is not supported by the env koanf.
func (p *Cliflag) ReadBytes() ([]byte, error) {
	return nil, errors.New("basicflag provider does not support this method")
}

// Watch is not supported.
func (p *Cliflag) Watch(cb func(event interface{}, err error)) error {
	return errors.New("basicflag provider does not support this method")
}
