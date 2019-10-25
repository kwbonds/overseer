package main

import (
	"fmt"
	k8seventwatcher "github.com/cmaster11/k8s-event-watcher"
	"github.com/cmaster11/overseer/test"
	"regexp"
	"strings"
)

type resultFilter struct {
	/*
		- type: 		type=k8s-event
		- tag: 			tag=my-k8s-cluster
		- input
		- target: 		target=10\.0\.123\.111
		- error:		error=(ssl|SSL)
		- isDedup:		isDedup=true/isDedup=false
		- recovered:	recovered=true/recovered=false
	*/

	Type      *k8seventwatcher.Regexp
	Tag       *k8seventwatcher.Regexp
	Input     *k8seventwatcher.Regexp
	Target    *k8seventwatcher.Regexp
	Error     *k8seventwatcher.Regexp
	IsDedup   *bool
	Recovered *bool
}

func (f *resultFilter) Matches(result *test.Result) bool {
	if f.Type != nil && !f.Type.MatchString(result.Type) {
		return false
	}
	if f.Tag != nil && !f.Tag.MatchString(result.Tag) {
		return false
	}
	if f.Input != nil && !f.Input.MatchString(result.Input) {
		return false
	}
	if f.Target != nil && !f.Target.MatchString(result.Target) {
		return false
	}
	if f.Error != nil && (result.Error == nil ||
		!f.Error.MatchString(*result.Error)) {
		return false
	}

	if f.IsDedup != nil && result.IsDedup != *f.IsDedup {
		return false
	}
	if f.Recovered != nil && result.Recovered != *f.Recovered {
		return false
	}

	return true
}

const commaTemporaryReplacement = "___COMMA_REPLACEMENT"

var regexpKeyQuery = regexp.MustCompile("^(\\w+)=(.*)$")

// Accepts a Filter query and returns a Filter object
//
// Filter query can be contain multiple options, divided by comma (,)
// For regex values, comma can be escaped with \,
func newResultFilterFromQuery(queryString string) (*resultFilter, error) {
	// Temporary replacement for comma
	queryString = strings.ReplaceAll(queryString, "\\,", commaTemporaryReplacement)

	// Split in all the different queries
	queries := strings.Split(queryString, ",")

	filter := &resultFilter{}

	for _, query := range queries {
		// Restore comma
		query = strings.ReplaceAll(query, commaTemporaryReplacement, ",")

		// Process query
		// Query has key=regex
		matches := regexpKeyQuery.FindStringSubmatch(query)
		if matches == nil {
			return nil, fmt.Errorf("invalid query: %s", query)
		}

		queryKey := matches[1]
		queryRegexString := matches[2]

		used := false

		switch queryKey {
		case "isDedup":
			used = true
			v := queryRegexString == "true"
			filter.IsDedup = &v
		case "recovered":
			used = true
			v := queryRegexString == "true"
			filter.Recovered = &v
			/*
				Type      *k8seventwatcher.Regexp
					Tag       *k8seventwatcher.Regexp
					Input     *k8seventwatcher.Regexp
					Target    *k8seventwatcher.Regexp
					Error     *k8seventwatcher.Regexp
			*/
		}

		if !used {
			queryRegex, err := k8seventwatcher.NewRegexp(queryRegexString)
			if err != nil {
				return nil, err
			}

			switch queryKey {
			case "type":
				filter.Type = queryRegex
			case "tag":
				filter.Tag = queryRegex
			case "input":
				filter.Input = queryRegex
			case "target":
				filter.Target = queryRegex
			case "error":
				filter.Error = queryRegex
			default:
				return nil, fmt.Errorf("unhandled Filter key: %s", queryKey)
			}
		}
	}

	return filter, nil
}
