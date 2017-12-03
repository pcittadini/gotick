package tasks

import (
	"strconv"
	"fmt"
	"os/signal"
	"os"
	"math/rand"
	"time"
	"github.com/Shopify/sarama"
	"github.com/satori/go.uuid"

	"database/sql"

	_ "github.com/lib/pq"
)

// Setup configuration
type Task struct {
	Name 		string 		`json:"name"`
	ConsumerID 	string 		`json:"consumerID"`
	Topic		string 		`json:"topic"`
	Specs       struct {
		RunEvery int 	`json:"runEvery"`
		Action   string `json:"action"`
	} `json:"specs"`
}

func (t *Task)Scheduler(){


	// check if task is already registered
	// if true; then add a new consumer to the task

	err := updateOrCreateTask(t)
	if err != nil {
		panic(err)
	}

	//err = updateOrCreateConsumer(t)
	//if err != nil {
	//	panic(err)
	//}

	config := sarama.NewConfig()

	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	brokers := []string{"localhost:9092"}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var enqueued, errors int
	doneCh := make(chan struct{})



	go func() {
		for {
			// mock select consumer (round-robin)
			//rand := random(1,3)
			time.Sleep(time.Duration(t.Specs.RunEvery) * time.Second)
			message := "{\"executor\":\"" + t.ConsumerID + "\", \"action\":\"testAction\"}"
			strTime := strconv.Itoa(int(time.Now().Unix()))
			msg := &sarama.ProducerMessage{
				Topic: t.Topic,
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder(message),
			}
			select {
			case producer.Input() <- msg:
				enqueued++
				fmt.Println("[producer-1] Produce message")
			case err := <-producer.Errors():
				errors++
				fmt.Println("[producer] Failed to produce message:", err)
			case <-signals:
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Printf("[producer] Enqueued: %d; errors: %d\n", enqueued, errors)
}

func updateOrCreateTask(t *Task)(err error){

	var found bool
	// TODO store in conf
	connStr := "postgres://postgres:postgres@10.0.75.1:32768/gotick?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// check if task already exist
	q, err := db.Query(`SELECT * from tasks WHERE name=$1`,t.Name)
	if err != nil {
		fmt.Println(err.Error())
	}


	for q.Next() {
		var uid string
		var name string
		var action string

		err = q.Scan(&uid, &name, &action)
		if err != nil {
			fmt.Printf(err.Error())
		}else{
			fmt.Println("found task in store")
			fmt.Println("uid | name | action ")
			fmt.Printf("%3v | %8v | %6v\n", uid, name, action)
			found = true
			// TODO - here check consumer, and add consumer if !exist
		}

	}
	// new task
	if !found {
		fmt.Println("new task, adding in store")
		uuid := uuid.NewV4()
		_ = db.QueryRow(`INSERT INTO tasks(id, name, action)VALUES($1,$2,$3)`,uuid , t.Name, t.Specs.Action)
	}

	// TODO - after task has been addedd add a consumer and reload the consumer group

	return
}

func updateOrCreateConsumer(t *Task)(err error){
	return nil
}


// mock of getWokers
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}