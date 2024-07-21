package kafkademo

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"log"
	"strconv"
	"time"
)

type Consumer struct {
	ready chan bool
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			log.Printf("Message claimed partition: %+v, offset: %+v, topic: %+v, key: %+v, value: %+v", msg.Partition, msg.Offset, msg.Topic, msg.Key, msg.Value)
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

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
		Topic:     topic,
		Key:       sarama.StringEncoder("test"),
		Timestamp: time.Now(),
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
				log.Printf("Send message failed, err: %+v", err)
				continue
			}
			log.Printf("Send msg: %+v", msg)
			value++

			time.Sleep(time.Second * 5)
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

// 消费者组
func KafkaConsumerGroup(group, topic string, addrs []string, stop chan struct{}) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	ctx, _ := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(addrs, group, config)
	if err != nil {
		log.Fatalf("New consumer group failed, err: %+v", err)
		return
	}

	consumer := Consumer{
		ready: make(chan bool),
	}

	go func() {
		for {
			select {
			case <-stop:
				log.Printf("Stop consume msg")
			default:
				if err = client.Consume(ctx, []string{topic}, &consumer); err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						return
					}
					log.Panicf("Error from consumer: %+v", err)
				}
				if ctx.Err() != nil {
					return
				}
				consumer.ready = make(chan bool)
			}
		}
	}()

	<-consumer.ready
	log.Println("Sarama consumer up and running...")

	isPaused := false
	toggle := func() {
		if isPaused {
			client.ResumeAll()
		} else {
			client.PauseAll()
		}
		isPaused = !isPaused
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			toggle()
		}
	}
}
