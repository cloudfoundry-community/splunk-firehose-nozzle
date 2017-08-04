package logging

//go:generate counterfeiter . Logging

type Logging interface {
	Open() error
	Close() error
	Log(fields map[string]interface{}, msg string) error
}
