package main

import (
	"fmt"
	"regexp"
)

var regexDestinationQueue = regexp.MustCompile("([\\w.-]+)(\\[(.*)])?")

type destinationQueue struct {
	queueKey string
	filter   *resultFilter
}

func newDestinationQueueFromString(value string) (*destinationQueue, error) {
	matches := regexDestinationQueue.FindStringSubmatch(value)
	if matches == nil {
		return nil, fmt.Errorf("invalid destination queue value: %s", value)
	}

	queue := &destinationQueue{
		queueKey: matches[1],
	}

	if len(matches) == 3 && matches[2] != "" {
		filtersString := matches[2]
		filter, err := newResultFilterFromQuery(filtersString)
		if err != nil {
			return nil, fmt.Errorf("invalid queue filter: %s", filtersString)
		}

		queue.filter = filter
	}

	return queue, nil
}
