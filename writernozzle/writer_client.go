package writernozzle

import (
	"io"
)

type WriterClient struct {
	writer io.Writer
}

func NewWriterClient(writer io.Writer) *WriterClient {
	return &WriterClient{
		writer: writer,
	}
}

func (w *WriterClient) PostBatch(events []interface{}) error {
	for _, event := range events {
		_, err := w.writer.Write(event.([]byte))
		if err != nil {
			return err
		}
	}
	return nil
}
