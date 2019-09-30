package gofherd

type Logger interface {
	Printf(format string, v ...interface{})
}

type NoOpLogger struct {
}

func (noop NoOpLogger) Printf(format string, v ...interface{}) {
}
