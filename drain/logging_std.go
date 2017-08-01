package drain

import "fmt"

type LoggingStd struct{}

func (l *LoggingStd) Connect() error {
	return nil
}

func (l *LoggingStd) Close() error {
	return nil
}

func (l *LoggingStd) ShipEvents(fields map[string]interface{}, msg string) error {
	fmt.Printf("%+v\n", fields)
	if len(msg) > 0 {
		fmt.Printf("\t%s\n", msg)
	}
	return nil
}
