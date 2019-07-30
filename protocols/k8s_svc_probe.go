// Kubernetes Service Tester
//
// The Kubernetes service tester checks that a k8s service has more than the specified number of endpoints (default >= 1).
//
// This test is invoked via input like so:
//
//    service-doman must run k8s-svc
//

package protocols

import (
	"fmt"
	"strconv"

	"github.com/skx/overseer/test"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8SSvcTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *K8SSvcTest) Arguments() map[string]string {
	known := map[string]string{
		"endpoints": "^[0-9]+$",
		"namespace": "^[a-z0-9]+(?:-[a-z0-9]+)*$",
	}
	return known
}

func (s *K8SSvcTest) ShouldResolveHostname() bool {
	return false
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *K8SSvcTest) Example() string {
	str := `
K8SSvc Tester
-------------
 The Kubernetes service tester checks that a k8s service has 
 more than the specified number of endpoints (default >= 1).

 It expects overseer worker to run from inside a k8s cluster.

 This test is invoked via input like so:

    service-name must run k8s-svc

 The number of min endpoints that need to be available can be set with:

	# Requires minimum 2 endpoints to be available for the test to succeed
	service-name must run k8s-svc with endpoints 2

 The namespace where to look for the service can be set with:

	service-name must run k8s-svc with namespace 'namespace-name'
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
func (s *K8SSvcTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	//
	// The default port to connect to.
	//
	minEndpoints := 1
	namespace := ""

	//
	// If the user specified a different port update to use it.
	//
	if tst.Arguments["endpoints"] != "" {
		minEndpoints, err = strconv.Atoi(tst.Arguments["endpoints"])
		if err != nil {
			return err
		}
	}
	if tst.Arguments["namespace"] != "" {
		namespace = tst.Arguments["endpoints"]
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(target, v1.GetOptions{})
	if err != nil {
		return err
	}

	// Count the number of available endpoints
	endpointsCount := 0

	for _, v := range endpoints.Subsets {
		endpointsCount += len(v.Addresses)
	}

	if endpointsCount < minEndpoints {
		return fmt.Errorf("number of available endpoints (%d) is lower than min defined (%d)", endpointsCount, minEndpoints)
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("k8s_svc", func() ProtocolTest {
		return &K8SSvcTest{}
	})
}
