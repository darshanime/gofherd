package gofherd

// Logger interface is accepted by SetLogger function and used to log the output.
// It has a single function Printf with the signature: `Printf(format string, v ...interface{})`
type Logger interface {
	Printf(format string, v ...interface{})
}

type noOpLogger struct {
}

func (noop noOpLogger) Printf(format string, v ...interface{}) {
}
