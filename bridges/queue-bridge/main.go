//
// This is the queue bridge, which should be built like so:
//
//     go build .
//
// Once built launch it as follows:
//
//     $ ./queue-bridge [-redis-queue-key=overseer.results] -dest-queue overseer.results.email -dest-queue overseer.results.webhook[tag=hello.*,name=dsf]
//
// It is possible to conditionally clone results to different queues by using regex filters, e.g.
//
// -destination-queues=overseer.results.webhook[tag=k8s-cluster.*]
//
// Results are filterable on:
//
// - type: 		type=k8s-event
// - tag: 		tag=my-k8s-cluster
// - input
// - target: 	target=10\.0\.123\.111
// - error:		error=(ssl|SSL)
// - isDedup:	isDedup=true/isDedup=false
// - recovered:	recovered=true/recovered=false
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
)

type QueueBridge struct {
	R *redis.Client

	// The queues to use as destination
	Queues []*destinationQueue
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

	for _, queue := range bridge.Queues {
		if queue.filter != nil && !queue.filter.Matches(testResult) {
			continue
		}

		_, err = bridge.R.RPush(queue.queueKey, msg).Result()
		if err != nil {
			fmt.Printf("Result clone failed for queue [%s]: %s\n", queue.queueKey, err)
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

	var queuesArray stringsFlag

	flag.Var(&queuesArray, "dest-queue", "The redis queues to clone results into")

	flag.Parse()

	var queues []*destinationQueue
	for _, queueString := range queuesArray {
		queue, err := newDestinationQueueFromString(queueString)
		if err != nil {
			fmt.Printf("invalid queue string: %+v\n", queueString)
			os.Exit(1)
		}

		queues = append(queues, queue)
	}

	//
	// Sanity-check.
	//
	if len(queues) == 0 {
		fmt.Printf("Usage: ./queue-bridge [-redis-queue-key=overseer.results] -dest-queue=overseer.results.queue -dest-queue=overseer.results.webhook\n")
		os.Exit(1)
	}

	fmt.Printf("started with %d queues", len(queues))

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
		Queues: queues,
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
