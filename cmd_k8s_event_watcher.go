// K8s event watcher
//
// The k8s-event-watcher sub-command monitors a k8s cluster events stream and triggers alerts when matching specific
// conditions.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cmaster11/k8s-event-watcher"
	"github.com/cmaster11/overseer/test"
	"github.com/go-redis/redis"
	"github.com/google/subcommands"
	"k8s.io/api/core/v1"
)

// This is our structure, largely populated by command-line arguments
type k8sEventWatcherCmd struct {
	// K8s configuration path, can be empty
	KubeConfigPath string

	// Events filter configuration path
	EventFilterConfigPath string

	// Default amount of events repetitions before triggering an error
	// MinRepetitions uint

	// Default deduplication duration
	// DedupDuration time.Duration

	// The redis-host we're going to connect to for our queues.
	RedisHost string

	// The redis-database we're going to use.
	RedisDB int

	// The (optional) redis-password we'll use.
	RedisPassword string

	// The redis-socket we're going to use. (If used, we ignore the specified host / port)
	RedisSocket string

	// Redis connection timeout
	RedisDialTimeout time.Duration

	// Tag applied to all results
	Tag string

	// Should the watcher be verbose?
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
// func (p *k8sEventWatcherCmd) verbose(txt string) {
//	if p.Verbose {
//		fmt.Print(txt)
//	}
// }

//
// Flag setup.
//
func (p *k8sEventWatcherCmd) SetFlags(f *flag.FlagSet) {

	//
	// Setup the default options here, these can be loaded/replaced
	// via a configuration-file if it is present.
	//
	var defaults k8sEventWatcherCmd
	// defaults.MinRepetitions = 0
	// defaults.DedupDuration = 0
	defaults.Tag = ""
	defaults.Verbose = false
	defaults.RedisHost = "localhost:6379"
	defaults.RedisDB = 0
	defaults.RedisPassword = ""
	defaults.RedisDialTimeout = 5 * time.Second
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
	// f.UintVar(&p.MinRepetitions, "min-repetitions", defaults.MinRepetitions, "How many times to an event has to occur before triggering an error.")

	// f.DurationVar(&p.DedupDuration, "dedup", defaults.DedupDuration, "The maximum duration of a deduplication.")

	// Redis
	f.StringVar(&p.RedisHost, "redis-host", defaults.RedisHost, "Specify the address of the redis queue.")
	f.IntVar(&p.RedisDB, "redis-db", defaults.RedisDB, "Specify the database-number for redis.")
	f.StringVar(&p.RedisPassword, "redis-pass", defaults.RedisPassword, "Specify the password for the redis queue.")
	f.StringVar(&p.RedisSocket, "redis-socket", defaults.RedisSocket, "If set, will be used for the redis connections.")
	f.DurationVar(&p.RedisDialTimeout, "redis-timeout", defaults.RedisDialTimeout, "Redis connection timeout.")

	// Tag
	f.StringVar(&p.Tag, "tag", defaults.Tag, "Specify the tag to add to all events.")
}

// notify is used to store the result of a test in our redis queue.
func (p *k8sEventWatcherCmd) onEvent(event *v1.Event, eventFilter *k8seventwatcher.EventFilter) {

	//
	// If we don't have a redis-server then return immediately.
	//
	// (This shouldn't happen, as without a redis-handle we can't
	// fetch jobs to execute.)
	//
	if p._r == nil {
		return
	}

	target := fmt.Sprintf("%s/%s/%s", event.InvolvedObject.Namespace, event.InvolvedObject.Kind, event.InvolvedObject.Name)
	input := fmt.Sprintf("%s [%s]", target, eventFilter.StringShort())

	testResult := &test.Result{
		Input:  input,
		Target: target,
		Time:   event.CreationTimestamp.Unix(),
		Type:   "k8s-event",
		Tag:    p.Tag,
	}

	eventFilterString := eventFilter.ToYAML()

	errorString := strings.TrimSpace(fmt.Sprintf(`
%s

- Event reason: %s
- Event type: %s
- Object namespace: %s
- Object kind: %s
- Object name: %s
- Event filter:
%s
`,
		event.Message,
		event.Reason,
		event.Type,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.Kind,
		event.InvolvedObject.Name,
		indent(eventFilterString, "    "),
	))
	testResult.Error = &errorString

	//
	// Convert the event to a JSON string we can notify.
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
			Network:     "unix",
			Addr:        p.RedisSocket,
			Password:    p.RedisPassword,
			DB:          p.RedisDB,
			DialTimeout: p.RedisDialTimeout,
		})
	} else {
		p._r = redis.NewClient(&redis.Options{
			Addr:        p.RedisHost,
			Password:    p.RedisPassword,
			DB:          p.RedisDB,
			DialTimeout: p.RedisDialTimeout,
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
	if err = eventWatcher.Start(p.onEvent); err != nil {
		fmt.Printf("K8s event watcher start failed: %s\n", err.Error())
		return subcommands.ExitFailure
	}

	defer eventWatcher.Stop()

	//
	// Wait for events, in a blocking-manner.
	//
	fmt.Println("Press 'CTRL-C' to exit...")
	waitForCtrlC()

	return subcommands.ExitSuccess
}
