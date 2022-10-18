package utils

import "fmt"

const (
	FloatType = CounterType("float64")
	UintType  = CounterType("uint64")
)

type CounterType string

type Counter interface {
	Add(interface{})
	Reset()
	Clone() Counter
	Value() interface{}
}

type IntCounter uint64

func (ic *IntCounter) Add(num interface{}) {
	// fmt.Println("insideeeee utilsssss int")
	switch n := num.(type) {
	case uint64:
		*ic = *ic + IntCounter(n)
	case IntCounter:
		*ic = *ic + n
	case int:
		*ic = *ic + IntCounter(n)
	case float64:
		*ic = *ic + IntCounter(n)
	default:
		fmt.Printf("%V - %V", num, n)
	}
}

func (ic *IntCounter) Clone() Counter {
	counter := new(IntCounter)
	counter.Add(*ic)
	return counter
}

func (ic *IntCounter) Value() interface{} {

	return uint64(*ic)
}

func (ic *IntCounter) Reset() {
	*ic = 0
}

type NopCounter struct{}

func (nc *NopCounter) Add(num interface{}) {

}

func (nc *NopCounter) Reset() {

}

func (nc *NopCounter) Clone() Counter {
	return &NopCounter{}
}

func (nc *NopCounter) Value() interface{} { return &NopCounter{} }
