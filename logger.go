package gofherd

type Logger interface {
	Printf(format string, v ...interface{})
}

type noOpLogger struct {
}

func (noop noOpLogger) Printf(format string, v ...interface{}) {
}
