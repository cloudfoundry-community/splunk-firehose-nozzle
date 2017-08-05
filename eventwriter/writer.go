package eventwriter

type Writer interface {
	Write([]map[string]interface{}) error
}
