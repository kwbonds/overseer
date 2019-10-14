// K8s event watcher
//
// The k8s-event-watcher sub-command monitors a k8s cluster events stream and triggers alerts when matching specific
// conditions.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	k8seventwatcher "github.com/cmaster11/k8s-event-watcher"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"os"
	"time"

	"github.com/cmaster11/overseer/test"
	"github.com/go-redis/redis"
	"github.com/google/subcommands"
	_ "github.com/skx/golang-metrics"
)

// This is our structure, largely populated by command-line arguments
type k8sEventWatcherCmd struct {
	// K8s configuration path, can be empty
	KubeConfigPath string

	// Events filter configuration path
	EventFilterConfigPath string

	// Default amount of events repetitions before triggering an error
	MinRepetitions uint

	// Default deduplication duration
	DedupDuration time.Duration

	// The redis-host we're going to connect to for our queues.
	RedisHost string

	// The redis-database we're going to use.
	RedisDB int

	// The (optional) redis-password we'll use.
	RedisPassword string

	// The redis-sockt we're going to use. (If used, we ignore the specified host / port)
	RedisSocket string

	// Tag applied to all results
	Tag string

	// Should the testing, and the tests, be verbose?
	Verbose bool

	// The handle to our redis-server
	_r *redis.Client
}

//
// Glue
//
func (*k8sEventWatcherCmd) Name() string { return "k8s-event-watcher" }
func (*k8sEventWatcherCmd) Synopsis() string {
	return "Watches for k8s events and triggers alerts when conditions are met"
}
func (*k8sEventWatcherCmd) Usage() string {
	return `k8s-event-watcher :
  Watches for k8s events and triggers alerts when conditions are met.
`
}

// verbose shows a message only if we're running verbosely
//func (p *k8sEventWatcherCmd) verbose(txt string) {
//	if p.Verbose {
//		fmt.Print(txt)
//	}
//}

//
// Flag setup.
//
func (p *k8sEventWatcherCmd) SetFlags(f *flag.FlagSet) {

	//
	// Setup the default options here, these can be loaded/replaced
	// via a configuration-file if it is present.
	//
	var defaults k8sEventWatcherCmd
	defaults.MinRepetitions = 0
	defaults.DedupDuration = 0
	defaults.Tag = ""
	defaults.Verbose = false
	defaults.RedisHost = "localhost:6379"
	defaults.RedisDB = 0
	defaults.RedisPassword = ""
	defaults.KubeConfigPath = ""
	defaults.EventFilterConfigPath = ""

	//
	// If we have a configuration file then load it
	//
	if len(os.Getenv("OVERSEER")) > 0 {
		cfg, err := ioutil.ReadFile(os.Getenv("OVERSEER"))
		if err == nil {
			err = json.Unmarshal(cfg, &defaults)
			if err != nil {
				fmt.Printf("WARNING: Error loading overseer.json - %s\n",
					err.Error())
			}
		} else {
			fmt.Printf("WARNING: Failed to read configuration-file - %s\n",
				err.Error())
		}
	}

	//
	// Allow these defaults to be changed by command-line flags
	//
	// Verbose
	f.BoolVar(&p.Verbose, "verbose", defaults.Verbose, "Show more output.")

	// Configuration
	f.StringVar(&p.KubeConfigPath, "kubeconfig", defaults.KubeConfigPath, "Kubernetes cluster configuration file, can be empty")
	f.StringVar(&p.EventFilterConfigPath, "watcher-config", defaults.EventFilterConfigPath, "Event watcher configuration file")

	// Retry
	f.UintVar(&p.MinRepetitions, "min-repetitions", defaults.MinRepetitions, "How many times to an event has to occur before triggering an error.")

	f.DurationVar(&p.DedupDuration, "dedup", defaults.DedupDuration, "The maximum duration of a deduplication.")

	// Redis
	f.StringVar(&p.RedisHost, "redis-host", defaults.RedisHost, "Specify the address of the redis queue.")
	f.IntVar(&p.RedisDB, "redis-db", defaults.RedisDB, "Specify the database-number for redis.")
	f.StringVar(&p.RedisPassword, "redis-pass", defaults.RedisPassword, "Specify the password for the redis queue.")
	f.StringVar(&p.RedisSocket, "redis-socket", defaults.RedisSocket, "If set, will be used for the redis connections.")

	// Tag
	f.StringVar(&p.Tag, "tag", defaults.Tag, "Specify the tag to add to all events.")
}

// notify is used to store the result of a test in our redis queue.
func (p *k8sEventWatcherCmd) onEvent(event *v1.Event, filterDescription string) {

	//
	// If we don't have a redis-server then return immediately.
	//
	// (This shouldn't happen, as without a redis-handle we can't
	// fetch jobs to execute.)
	//
	if p._r == nil {
		return
	}

	testResult := &test.Result{
		Input:  filterDescription,
		Target: filterDescription,
		Time:   event.CreationTimestamp.Unix(),
		Type:   "k8s-event",
		Tag:    p.Tag,
	}

	errorString := fmt.Sprintf(`
Event created: %s

- Event reason: %s
- Event type: %s
- Object namespace: %s
- Object kind: %s
- Object name: %s
`,
		event.Message,
		event.Reason,
		event.Type,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.Kind,
		event.InvolvedObject.Name,
	)
	testResult.Error = &errorString

	//if testDefinition.DedupDuration == nil && p.DedupDuration > 0 {
	//	// Assign a default dedup duration
	//	testDefinition.DedupDuration = &p.DedupDuration
	//}
	//
	//// If test has a deduplication rule, avoid re-triggering a notification if not needed, or clean the dedup cache if needed
	//if testDefinition.DedupDuration != nil {
	//
	//	hash := testResult.Hash()
	//	if testResult.Error != nil {
	//
	//		// Save the current notification time, this keeps alive the deduplication. *10 so that it's not going to expire
	//		// anytime soon.
	//		p.setDeduplicationCacheTime(hash, *testDefinition.DedupDuration*10)
	//
	//		lastAlertTime := p.getDeduplicationLastAlertTime(hash)
	//
	//		// With dedup, we don't want to trigger same notification, unless we just passed the dedup duration
	//		if lastAlertTime != nil {
	//			now := time.Now().Unix()
	//			diffLastAlert := now - *lastAlertTime
	//			dedupDurationSeconds := int64(*testDefinition.DedupDuration / time.Second)
	//
	//			if diffLastAlert < dedupDurationSeconds {
	//				// There is no need to trigger the notification, because not enough time has passed since the last one
	//				p.verbose(fmt.Sprintf("Skipping notification (dedup, last notif %s ago) for test `%s` (%s)\n",
	//					time.Duration(diffLastAlert)*time.Second,
	//					testDefinition.Input, testDefinition.Target))
	//				return nil
	//			}
	//
	//			// Let the user know that the generated notification is a duplicate
	//			testResult.IsDedup = true
	//		}
	//
	//		p.setDeduplicationLastAlertTime(hash, *testDefinition.DedupDuration*10)
	//
	//	} else {
	//		// Check if a dedup was happening
	//		dedupCacheTime := p.getDeduplicationCacheTime(hash)
	//
	//		// If there was a dedup cache time, we can mark this test as recovered
	//		if dedupCacheTime != nil {
	//			// Clear any dedup cache, because the test has passed
	//			p.clearDeduplicationCacheTime(hash)
	//			p.clearDeduplicationLastAlertTime(hash)
	//			testResult.Recovered = true
	//
	//			p.verbose(fmt.Sprintf("Test recovered: `%s` (%s)\n",
	//				testDefinition.Input, testDefinition.Target))
	//		}
	//
	//	}
	//
	//}

	//
	// Convert the test result to a JSON string we can notify.
	//
	j, err := json.Marshal(testResult)
	if err != nil {
		fmt.Printf("Failed to encode test-result to JSON: %s", err.Error())
		return
	}

	//
	// Publish the message to the queue.
	//
	_, err = p._r.RPush("overseer.results", j).Result()
	if err != nil {
		fmt.Printf("Result addition failed: %s\n", err)
		return
	}
}

//func (p *k8sEventWatcherCmd) getDeduplicationCacheKey(hash string) string {
//	return fmt.Sprintf("overseer.dedup-cache.%s", hash)
//}
//
//// TODO
//// TODO
//// TODO
//// TODO: share deduplication code
//// TODO
//// TODO
//
//func (p *k8sEventWatcherCmd) getDeduplicationCacheTime(hash string) *int64 {
//	if p._r == nil {
//		return nil
//	}
//
//	cacheKey := p.getDeduplicationCacheKey(hash)
//	cacheTime, err := p._r.Get(cacheKey).Int64()
//	if err != nil {
//		if err == redis.Nil {
//			// Key just does not exist
//			return nil
//		}
//
//		fmt.Printf("Failed to get dedup cache key: %s\n", err)
//		return nil
//	}
//
//	return &cacheTime
//}
//
//func (p *k8sEventWatcherCmd) setDeduplicationCacheTime(hash string, expiry time.Duration) {
//	if p._r == nil {
//		return
//	}
//
//	cacheKey := p.getDeduplicationCacheKey(hash)
//	_, err := p._r.Set(cacheKey, time.Now().Unix(), expiry).Result()
//	if err != nil {
//		fmt.Printf("Failed to set dedup cache key: %s\n", err)
//	}
//}
//
//func (p *k8sEventWatcherCmd) clearDeduplicationCacheTime(hash string) {
//	if p._r == nil {
//		return
//	}
//
//	cacheKey := p.getDeduplicationCacheKey(hash)
//	_, err := p._r.Del(cacheKey).Result()
//	if err != nil {
//		fmt.Printf("Failed to clear dedup cache key: %s\n", err)
//	}
//}
//
//func (p *k8sEventWatcherCmd) getDeduplicationLastAlertKey(hash string) string {
//	return fmt.Sprintf("overseer.dedup-last-alert.%s", hash)
//}
//
//func (p *k8sEventWatcherCmd) getDeduplicationLastAlertTime(hash string) *int64 {
//	if p._r == nil {
//		return nil
//	}
//
//	cacheKey := p.getDeduplicationLastAlertKey(hash)
//	cacheTime, err := p._r.Get(cacheKey).Int64()
//	if err != nil {
//		if err == redis.Nil {
//			// Key just does not exist
//			return nil
//		}
//
//		fmt.Printf("Failed to get dedup last alert key: %s\n", err)
//		return nil
//	}
//
//	return &cacheTime
//}
//
//func (p *k8sEventWatcherCmd) setDeduplicationLastAlertTime(hash string, expiry time.Duration) {
//	if p._r == nil {
//		return
//	}
//
//	cacheKey := p.getDeduplicationLastAlertKey(hash)
//	_, err := p._r.Set(cacheKey, time.Now().Unix(), expiry).Result()
//	if err != nil {
//		fmt.Printf("Failed to set dedup last alert key: %s\n", err)
//	}
//}
//
//func (p *k8sEventWatcherCmd) clearDeduplicationLastAlertTime(hash string) {
//	if p._r == nil {
//		return
//	}
//
//	cacheKey := p.getDeduplicationLastAlertKey(hash)
//	_, err := p._r.Del(cacheKey).Result()
//	if err != nil {
//		fmt.Printf("Failed to clear dedup last alert key: %s\n", err)
//	}
//}

// alphaNumeric removes all non alpha-numeric characters from the
// given string, and returns it.  We replace the characters that
// are invalid with `_`.
//func (p *k8sEventWatcherCmd) alphaNumeric(input string) string {
//	//
//	// Remove non alphanumeric
//	//
//	reg, err := regexp.Compile("[^A-Za-z0-9]+")
//	if err != nil {
//		panic(err)
//	}
//	return reg.ReplaceAllString(input, "_")
//}

//
// Entry-point.
//
func (p *k8sEventWatcherCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if p.EventFilterConfigPath == "" {
		fmt.Printf("Missing event watcher configuration\n")
		return subcommands.ExitFailure
	}

	//
	// Connect to the redis-host.
	//
	if p.RedisSocket != "" {
		p._r = redis.NewClient(&redis.Options{
			Network:  "unix",
			Addr:     p.RedisSocket,
			Password: p.RedisPassword,
			DB:       p.RedisDB,
		})
	} else {
		p._r = redis.NewClient(&redis.Options{
			Addr:     p.RedisHost,
			Password: p.RedisPassword,
			DB:       p.RedisDB,
		})
	}

	//
	// And run a ping, just to make sure it worked.
	//
	_, err := p._r.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		return subcommands.ExitFailure
	}

	//
	// Setup our the event watcher
	//
	var kubeConfigPath *string
	if p.KubeConfigPath != "" {
		kubeConfigPath = &p.KubeConfigPath
	}
	eventWatcher, err := k8seventwatcher.NewK8sEventWatcher(
		p.EventFilterConfigPath,
		kubeConfigPath,
		os.Stdout,
	)
	if err != nil {
		fmt.Printf("K8s event watcher setup failed: %s\n", err.Error())
		return subcommands.ExitFailure
	}

	if p.Verbose {
		eventWatcher.Debug = true
	}

	fmt.Printf("k8s event watcher worker started [tag=%s]\n", p.Tag)

	// Wait for k8s events
	if err := eventWatcher.Start(p.onEvent); err != nil {
		fmt.Printf("K8s event watcher start failed: %s\n", err.Error())
		return subcommands.ExitFailure
	}

	defer eventWatcher.Stop()

	//
	// Wait for jobs, in a blocking-manner.
	//
	fmt.Println("Press 'Enter' to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	return subcommands.ExitSuccess
}
