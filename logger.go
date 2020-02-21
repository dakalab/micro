package micro

// Logger is logger interface
type Logger interface {
	Printf(string, ...interface{})
}

// LoggerFunc is a bridge between Logger and any third party logger
type LoggerFunc func(string, ...interface{})

// Printf implements Logger interface
func (f LoggerFunc) Printf(msg string, args ...interface{}) { f(msg, args...) }

// dummy logger writes nothing
var dummyLogger = LoggerFunc(func(string, ...interface{}) {})
