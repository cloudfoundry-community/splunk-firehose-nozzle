package utils

const (
	FloatType = CounterType("float64")
	UintType  = CounterType("uint64")
)

type CounterType string

type Counter interface {
	Add(interface{})
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

type NopCounter struct{}

func (nc *NopCounter) Add(num interface{}) {

}

func (nc *NopCounter) Clone() Counter {
	return &NopCounter{}
}

func (nc *NopCounter) Value() interface{} { return &NopCounter{} }
