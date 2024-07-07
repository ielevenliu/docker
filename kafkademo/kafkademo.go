package kafkademo

import (
	"github.com/IBM/sarama"
	"log"
	"strconv"
	"time"
)

// 生产者
func KafkaProducer(topic string, addrs []string, stop chan struct{}) {
	// producer 配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	// 连接 kafka Topic
	client, err := sarama.NewSyncProducer(addrs, config)
	if err != nil {
		log.Fatalf("New producer failed, err: %+v", err)
		return
	}
	defer client.Close()

	// 构造 kafka 消息
	msg := &sarama.ProducerMessage{
		Topic: topic,
	}

	var value int64
	for {
		select {
		case <-stop:
			log.Printf("Stop send msg")
			return
		default:
			msg.Value = sarama.StringEncoder(strconv.FormatInt(value, 10))
			if err = client.SendMessages([]*sarama.ProducerMessage{msg}); err != nil {
				log.Fatalf(" Send message failed, err: %+v", err)
				continue
			}
			log.Printf("Send msg: %+v", msg)
			value++

			time.Sleep(time.Second)
		}
	}
}

// 消费者
func KafkaConsumer(topic string, addrs []string, stop chan struct{}) {
	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		log.Fatalf("New consumer failed, err: %+v", err)
		return
	}

	var partitions []int32
	if partitions, err = consumer.Partitions(topic); err != nil {
		log.Fatalf("Get partitions list failed, err: %+v", err)
		return
	}
	log.Printf("Partitions: %+v", partitions)

	// 遍历所有分区
	for p := range partitions {
		var pc sarama.PartitionConsumer
		if pc, err = consumer.ConsumePartition(topic, int32(p), sarama.OffsetNewest); err != nil {
			log.Fatalf("Failed to start consumer for partition %d, err: %+v", p, err)
			return
		}
		defer pc.AsyncClose()

		go func() {
			for msg := range pc.Messages() {
				log.Printf("Partition: %d, Offset: %d, Key: %+v, Value: %+v", msg.Partition, msg.Offset, msg.Key, msg.Value)
			}
		}()
	}
	for {
		select {
		case <-stop:
			log.Printf("Stop consume")
			return
		default:
			log.Printf("Consume")
			time.Sleep(time.Second)
		}
	}
}
