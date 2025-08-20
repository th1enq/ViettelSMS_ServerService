package mq

type Message struct {
	Key     string
	Headers map[string]string
	Body    []byte
	Topic   string
}
