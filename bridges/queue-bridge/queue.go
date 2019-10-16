package main

import (
	"fmt"
	"regexp"
	"strings"
)

var regexDestinationQueue = regexp.MustCompile("^([\\w.-]+)(\\[(.+)])?$")

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

	if len(matches) == 3 {
		filtersString := strings.TrimSpace(matches[2])

		if filtersString == "" {
			return nil, fmt.Errorf("empty filter tag: %s", value)
		}

		filter, err := newResultFilterFromQuery(filtersString)
		if err != nil {
			return nil, fmt.Errorf("invalid queue filter: %s", filtersString)
		}

		queue.filter = filter
	}

	return queue, nil
}
