package main

import (
	"fmt"
	"regexp"
	"strings"
)

var regexDestinationQueue = regexp.MustCompile(`^([\w.-]+)(?:\[(.+)])?$`)

type destinationQueue struct {
	QueueKey string
	Filter   *resultFilter
}

func newDestinationQueuesFromStringArray(queuesStringArray []string) ([]*destinationQueue, error) {
	var queues []*destinationQueue
	for _, queueString := range queuesStringArray {
		queue, err := newDestinationQueueFromString(queueString)
		if err != nil {
			return nil, fmt.Errorf("invalid queue string: %+v, %s", queueString, err)
		}

		queues = append(queues, queue)
	}

	return queues, nil
}

func newDestinationQueueFromString(value string) (*destinationQueue, error) {
	matches := regexDestinationQueue.FindStringSubmatch(value)
	if matches == nil {
		return nil, fmt.Errorf("invalid destination queue value: %s", value)
	}

	queue := &destinationQueue{
		QueueKey: matches[1],
	}

	if len(matches) == 3 {
		filtersString := strings.TrimSpace(matches[2])

		if filtersString == "" {
			return queue, nil
		}

		filter, err := newResultFilterFromQuery(filtersString)
		if err != nil {
			return nil, fmt.Errorf("invalid queue filter: %s, %s", filtersString, err)
		}

		queue.Filter = filter
	}

	return queue, nil
}
