package main

import (
	"github.com/pcittadini/gotick/lib/clients"
	"github.com/pcittadini/gotick/lib/restclient"
	"fmt"
)
func main() {

	// CREATE TASK (via HTTP)
	ep := new(restclient.Endpoint)
	ep.Host = "http://localhost:8080"
	ep.Path = "/tasks/new"
	ep.Method = "POST"

	payload := `{
		"name":"myClientTask0",
		"consumerId":"1",
		"topic":"gotick_tasks",
		"specs":{
			"runEvery":5,
			"action":"clearDataBaseEntries"
		}
	}`

	ep.Body = payload
	r, err := ep.Do()
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Result)
	fmt.Print("starting ...")

	// START TASK HANDLER (via BROKER)
	clients.TestConsumer(1)

	return
}