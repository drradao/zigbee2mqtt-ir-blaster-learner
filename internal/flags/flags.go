// Package flags defines the CLI flag values passed through the FX dependency
// injection container. It lives in its own package to avoid circular imports
// between the cmd layer and internal packages.
package flags

// FlagValues holds all parsed CLI flag values supplied to FX as a value object.
type FlagValues struct {
	ConfigPath   string
	MQTTHost     string
	MQTTPort     int
	MQTTUser     string
	MQTTPassword string
	Device       string
	LogFile      string
	Session      string
	// Changed tracks which flags were explicitly set on the command line so
	// that flag overrides only apply when the user actually passed the flag.
	Changed map[string]bool
}
