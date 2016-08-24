package splunk

import "fmt"

type LoggingStd struct{}

func (l *LoggingStd) Connect() bool {
	return true
}

func (l *LoggingStd) ShipEvents(fields map[string]interface{}, msg string) {
	fmt.Printf("%+v\n", fields)
	if len(msg) > 0 {
		fmt.Printf("\t%s\n", msg)
	}
}
