package analyzer

import "github.com/asiffer/gospot"

// SpotConfig is a simplified configuration for the
// Spot algorithm
type SpotConfig struct {
	Q         float64
	Low       bool
	Level     float64
	MaxExcess uint64
}

type SpotError struct {
	Err   error
	State *gospot.Spot
}

// DefaultSpotConfig is the default SpotConfig
var DefaultSpotConfig = NewSpotConfig()

// NewSpotConfig returns a default Spot config
func NewSpotConfig() *SpotConfig {
	return &SpotConfig{
		Q:         5e-4,
		Level:     0.99,
		Low:       false,
		MaxExcess: 250,
	}
}

type SpotOption func(*SpotConfig)

// WithQ sets the detection threshold (anomalies will be the events with
// probability lower than q)
func WithQ(q float64) SpotOption {
	return func(c *SpotConfig) {
		c.Q = q
	}
}

// WithLevel sets the tail level between 0 and 1
// This is a quantile close to 1 in practice
func WithLevel(level float64) SpotOption {
	return func(c *SpotConfig) {
		c.Level = level
	}
}

// WithLow configures the detection for high peaks or low peaks
func WithLow(low bool) SpotOption {
	return func(c *SpotConfig) {
		c.Low = low
	}
}

// WithMaxExcess set the size of the inner buffer to fit the data
// The higher, the lower variance
func WithMaxExcess(maxExcess uint64) SpotOption {
	return func(c *SpotConfig) {
		c.MaxExcess = maxExcess
	}
}

// NewSpot creates a new Spot instance for the default config
// with variadic options
func NewSpot(options ...SpotOption) *gospot.Spot {
	config := NewSpotConfig()
	for _, opt := range options {
		opt(config)
	}

	spot, err := gospot.NewSpot(
		config.Q,
		config.Low,
		true,
		config.Level,
		config.MaxExcess)
	if err != nil {
		panic(err)
	}

	return spot
}
