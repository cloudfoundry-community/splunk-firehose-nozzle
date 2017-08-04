package eventsink

import "fmt"

type Std struct{}

func (l *Std) Open() error {
	return nil
}

func (l *Std) Close() error {
	return nil
}

func (l *Std) Write(fields map[string]interface{}, msg string) error {
	fmt.Printf("%+v\n", fields)
	if len(msg) > 0 {
		fmt.Printf("\t%s\n", msg)
	}
	return nil
}
