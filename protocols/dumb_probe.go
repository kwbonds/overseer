// Dumb probe, for internal Overseer testing

package protocols

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/cmaster11/overseer/test"
	// Import all auth methods k8s
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// DumbTest is our object.
type DumbTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *DumbTest) Arguments() map[string]string {
	known := map[string]string{
		"duration-min": `^[-+]?([0-9]*(\.[0-9]*)?[a-z]+)+$`,
		"duration-max": `^[-+]?([0-9]*(\.[0-9]*)?[a-z]+)+$`,
	}
	return known
}

// ShouldResolveHostname returns if this protocol requires the hostname resolution of the first test argument
func (s *DumbTest) ShouldResolveHostname() bool {
	return false
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *DumbTest) Example() string {
	str := `
Dumb Tester
-------------
Performs a test of random duration and result.

	fake-name must run dumb-test with duration-min 2s with duration-max 10s
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
func (s *DumbTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	durationMin := 1 * time.Second
	durationMax := 5 * time.Second

	if tst.Arguments["duration-min"] != "" {
		durationMin, err = time.ParseDuration(tst.Arguments["duration-min"])
		if err != nil {
			return err
		}
	}
	if tst.Arguments["duration-max"] != "" {
		durationMax, err = time.ParseDuration(tst.Arguments["duration-max"])
		if err != nil {
			return err
		}
	}

	if durationMax < durationMin {
		return errors.New("duration-max must be > than duration-min")
	}

	fail := rand.Float64() >= 0.5
	waitFor := time.Duration(rand.Int63n(int64(durationMax-durationMin)) + int64(durationMin))

	time.Sleep(waitFor)

	if fail {
		return fmt.Errorf("dumb test failed (duration %s)", waitFor.String())
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("dumb-test", func() ProtocolTest {
		return &DumbTest{}
	})
}
