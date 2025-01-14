package log

type Logger interface {
	Debugf(prefix, format string, args ...interface{})
	Infof(prefix, format string, args ...interface{})
	Warnf(prefix, format string, args ...interface{})
	Errorf(prefix, format string, args ...interface{})
}

type Loggerp interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type loggerp struct {
	logger Logger
	prefix string
}

func CreatePrefixedLogger(prefix string, l Logger) Loggerp {
	return &loggerp{
		logger: l,
		prefix: prefix,
	}
}

func (l *loggerp) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(l.prefix, format, args...)
}

func (l *loggerp) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(l.prefix, format, args...)
}

func (l *loggerp) Infof(format string, args ...interface{}) {
	l.logger.Infof(l.prefix, format, args...)
}

func (l *loggerp) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(l.prefix, format, args...)
}
