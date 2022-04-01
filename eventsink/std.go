package eventsink

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type Std struct{}

func (l *Std) Open() error {
	return nil
}

func (l *Std) Close() error {
	return nil
}

func (l *Std) Write(fields *events.Envelope) error {
	fmt.Printf("%+v\n", fields)
	return nil
}
