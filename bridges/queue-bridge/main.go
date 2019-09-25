//
// This is the queue bridge, which should be built like so:
//
//     go build .
//
// Once built launch it as follows:
//
//     $ ./queue-bridge [-redis-queue-key=overseer.results] -destination-queues=overseer.results.email,overseer.results.webhook
//
// When a test is provided on the source queue, it gets cloned into the destination queues.
// This helps using multiple bridges, e.g. to send an queue and a webhook for each test result.
//
// When the queue bridge is used, the email and webhook bridges can be started like:
//
// 	   $ ./email-bridge -email=sysadmin@example.com,hello@gmail.com -redis-queue-key overseer.results.email
// 	   $ ./webhook-bridge -url=https://example.com/bla -redis-queue-key overseer.results.webhook
//
// Alberto
// --
//

package main

import (
	"flag"
	"fmt"
	"github.com/cmaster11/overseer/test"
	"github.com/go-redis/redis"
	"os"
	"strings"
)

type QueueBridge struct {
	R *redis.Client

	// The queues to use as destination
	Queues []string
}

//
// Given a JSON string decode it and post it via queue if it describes
// a test-failure.
//
func (bridge *QueueBridge) Process(msg []byte) {
	testResult, err := test.ResultFromJSON(msg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processing result: %+v\n", testResult)

	for _, queueKey := range bridge.Queues {
		_, err = bridge.R.RPush(queueKey, msg).Result()
		if err != nil {
			fmt.Printf("Result clone failed for queue [%s]: %s\n", queueKey, err)
		}
	}
}

//
// Entry Point
//
func main() {

	//
	// Parse our flags
	//
	redisHost := flag.String("redis-host", "127.0.0.1:6379", "Specify the address of the redis queue.")
	redisPass := flag.String("redis-pass", "", "Specify the password of the redis queue.")
	redisQueueKey := flag.String("redis-queue-key", "overseer.results", "Specify the redis queue key to use as source.")

	queuesStr := flag.String("destination-queues", "", "The redis queues to clone results into")

	flag.Parse()

	queuesSplit := strings.Split(*queuesStr, ",")
	var queuesValid []string
	for _, queue := range queuesSplit {
		if queue == "" {
			continue
		}

		queuesValid = append(queuesValid, queue)
	}

	//
	// Sanity-check.
	//
	if len(queuesValid) == 0 {
		fmt.Printf("Usage: ./queue-bridge [-redis-queue-key=overseer.results] -destination-queues=overseer.results.queue,overseer.results.webhook\n")
		os.Exit(1)
	}

	//
	// Create the redis client
	//
	r := redis.NewClient(&redis.Options{
		Addr:     *redisHost,
		Password: *redisPass,
		DB:       0, // use default DB
	})

	//
	// And run a ping, just to make sure it worked.
	//
	_, err := r.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		os.Exit(1)
	}

	bridge := QueueBridge{
		R:      r,
		Queues: queuesValid,
	}

	for {

		//
		// Get test-results
		//
		msg, _ := r.BLPop(0, *redisQueueKey).Result()

		//
		// If they were non-empty, process them.
		//
		//   msg[0] will be "overseer.results"
		//
		//   msg[1] will be the value removed from the list.
		//
		if len(msg) >= 1 {
			bridge.Process([]byte(msg[1]))
		}
	}
}
