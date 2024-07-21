package kafkademo

import "testing"

func TestKafkaProducer(t *testing.T) {
	KafkaProducer("demo-topic-1", []string{"192.168.1.5:9092"}, make(chan struct{}))
}

func TestKafkaConsumer(t *testing.T) {
	KafkaConsumer("demo-topic-1", []string{"192.168.1.5:9092"}, make(chan struct{}))
}
