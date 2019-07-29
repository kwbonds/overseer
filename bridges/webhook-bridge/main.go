//
// This is the webhook bridge, which should be built like so:
//
//     go build .
//
// Once built launch it as follows:
//
//     $ ./webhook-bridge -url=https://example.com/bla
//
// When a test fails a webhook will sent
//
// Alberto
// --
//

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/skx/overseer/test"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// The url we notify
var webhookURL *string

// The redis handle
var r *redis.Client

// The redis connection details
var redisHost *string
var redisPass *string

//
// Given a JSON string decode it and post it via webhook if it describes
// a test-failure.
//
func process(msg []byte) {
	testResult := new(test.Result)

	if err := json.Unmarshal(msg, testResult); err != nil {
		panic(err)
	}

	//
	// If the test passed then we don't care.
	//
	if testResult.Error == nil {
		return
	}

	res, err := http.Post(*webhookURL, "application/json", bytes.NewBuffer(msg))
	if err != nil {
		fmt.Printf("Failed to execute webhook request: %s\n", err.Error())
		return
	}

	//
	// OK now we've submitted the post.
	//
	// We should retrieve the status-code + body, if the status-code
	// is "odd" then we'll show them.
	//
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading response to post: %s\n", err.Error())
		return
	}
	status := res.StatusCode

	if status < 200 || status >= 400 {
		fmt.Printf("Error - Status code was not successful: %d\n", status)
		fmt.Printf("Response - %s\n", body)
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
	webhookURL = flag.String("url", "", "The url address to notify")
	flag.Parse()

	//
	// Sanity-check.
	//
	if *webhookURL == "" {
		fmt.Printf("Usage: webhook-bridge -url=https://example.com/bla [-redis-host=127.0.0.1:6379] [-redis-pass=foo]\n")
		os.Exit(1)
	}

	_, err := url.Parse(*webhookURL)
	if err != nil {
		fmt.Printf("Failed to parse provided URL: %s\n", err.Error())
		os.Exit(1)
	}

	//
	// Create the redis client
	//
	r = redis.NewClient(&redis.Options{
		Addr:     *redisHost,
		Password: *redisPass,
		DB:       0, // use default DB
	})

	//
	// And run a ping, just to make sure it worked.
	//
	_, err = r.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		os.Exit(1)
	}

	for true {

		//
		// Get test-results
		//
		msg, _ := r.BLPop(0, "overseer.results").Result()

		//
		// If they were non-empty, process them.
		//
		//   msg[0] will be "overseer.results"
		//
		//   msg[1] will be the value removed from the list.
		//
		if len(msg) >= 1 {
			process([]byte(msg[1]))
		}
	}
}
