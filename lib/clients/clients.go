package clients

import (
	"fmt"
	"gopkg.in/Shopify/sarama.v1"
	"os"
	"os/signal"
)

func TestConsumer(consumerId int) {
	fmt.Printf("[consumer-%d] started", consumerId)
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Specify brokers address. This is default one
	brokers := []string{"localhost:9092"}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "gotick_tasks"
	// How to decide partition, is it fixed value...?
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("[consumer-%d] Received messages: %s %s \n", consumerId, string(msg.Key), string(msg.Value))
			case <-signals:
				fmt.Printf("[consumer-%d] Interrupt is detected \n", consumerId)
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Printf("[consumer-%d] Processed %d messages\n", consumerId, msgCount)
}
