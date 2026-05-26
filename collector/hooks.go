package collector

import "github.com/asiffer/netspot/register"

type Hook interface {
	OnData(hook func(*Data)) Hook
	OnError(hook func(error)) Hook
	OnLog(hook func(string)) Hook
	OnEnd(hook func(error)) Hook
}

// CollectorHooks implements Hook interface
type CollectorHooks struct {
	dataHooks  *register.Register[*Data]
	errorHooks *register.Register[error]
	logHooks   *register.Register[string]
	endHooks   *register.Register[error]
}

func NewCollectorHooks() CollectorHooks {
	return CollectorHooks{
		dataHooks:  register.NewRegister[*Data](),
		errorHooks: register.NewRegister[error](),
		logHooks:   register.NewRegister[string](),
		endHooks:   register.NewRegister[error](),
	}
}

// OnData appends a new callback to the data hook
func (c *CollectorHooks) OnData(hook func(*Data)) Hook {
	c.dataHooks.Register(hook)
	return c
}

// OnError appends a new callback to the error hook
func (c *CollectorHooks) OnError(hook func(error)) Hook {
	c.errorHooks.Register(hook)
	return c
}

// OnLog appends a new callback to the logging hook
func (c *CollectorHooks) OnLog(hook func(string)) Hook {
	c.logHooks.Register(hook)
	return c
}

// OnEnd appends a new callback to the end hook
// This is used to signal the end of the collector
func (c *CollectorHooks) OnEnd(hook func(error)) Hook {
	c.endHooks.Register(hook)
	return c
}

// Send sends data to registered hooks
func (c *CollectorHooks) send(data *Data) {
	c.dataHooks.Exec(data)
}

// Err sends error to registered hooks
func (c *CollectorHooks) err(err error) {
	c.errorHooks.Exec(err)
}

// Log sends logs to registered hooks
func (c *CollectorHooks) log(msg string) {
	c.logHooks.Exec(msg)
}

// End sends end signal to registered hooks
// This is used to signal the end of the collector
func (c *CollectorHooks) end(err error) {
	c.endHooks.Exec(err)
}
