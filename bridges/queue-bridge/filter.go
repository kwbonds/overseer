package main

import (
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

	Type      *regexp.Regexp
	Tag       *regexp.Regexp
	Input     *regexp.Regexp
	Target    *regexp.Regexp
	Error     *regexp.Regexp
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

// Accepts a filter query and returns a filter object
//
// Filter query can be contain multiple options, divided by comma (,)
// For regex values, comma can be escaped with \,
func NewResultFilterFromQuery(queryString string) (*resultFilter, error) {
	// Temporary replacement for comma
	queryString = strings.ReplaceAll(queryString, "\\,", commaTemporaryReplacement)

	// Split in all the different queries
	queries := strings.Split(queryString, ",")

	for _, query := range queries {
		// Restore comma
		query = strings.ReplaceAll(query, commaTemporaryReplacement, ",")

		// Process query
		// TODO
	}
}
