package main

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestNewDestinationQueueFromString(t *testing.T) {

	testSyntaxOK := func(t *testing.T, queryStringArray []string) {
		queues, err := newDestinationQueuesFromStringArray(queryStringArray)
		if err != nil {
			t.Fatalf("bad query: %+v, %s", queryStringArray, err)
		}
		m, _ := yaml.Marshal(queues)
		t.Logf("queries: %+v, queues:\n%s", queryStringArray, string(m))
	}
	testSyntaxBad := func(t *testing.T, queryStringArray []string) {
		_, err := newDestinationQueuesFromStringArray(queryStringArray)
		if err == nil {
			t.Fatalf("should have been bad query: %+v", queryStringArray)
		}
	}

	testSyntaxOK(t, []string{"query1[tag=asd.*]"})
	testSyntaxOK(t, []string{"query2"})
	testSyntaxOK(t, []string{"hello[isDedup=true,error=a.*]]"})
	testSyntaxOK(t, []string{"query1[tag=asd.*]", "query2", "hello[isDedup=true,error=a.*]]"})
	testSyntaxOK(t, []string{"overseer.results.webhook[tag=my-cluster-.*]"})

	testSyntaxBad(t, []string{"a[]", "b"})
}
