package kafkademo

import "testing"

func TestKafkaProducer(t *testing.T) {
	KafkaProducer("kafka-test", []string{"192.168.1.6:9092"}, make(chan struct{}))
}

func TestKafkaConsumer(t *testing.T) {
	KafkaConsumer("kafka-test", []string{"192.168.1.6:9092"}, make(chan struct{}))
}
