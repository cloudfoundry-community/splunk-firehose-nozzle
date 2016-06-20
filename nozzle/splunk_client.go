package nozzle

type SplunkEvent struct {
	Time       float64 `json:"time,omitempty"`
	Host       string  `json:"host,omitempty"`
	Source     string  `json:"source,omitempty"`
	SourceType string  `json:"sourcetype,omitempty"`
	Index      string  `json:"index,omitempty"`

	Event interface{} `json:"event"`
}

type SplunkClient interface {
	Post(SplunkEvent) error
}

type splunkClient struct {
}

func NewSplunkClient() SplunkClient {
	return &splunkClient{
	}
}

func (s *splunkClient) Post(event SplunkEvent) error {
	panic("tdd me")
}
