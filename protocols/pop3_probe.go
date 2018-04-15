package protocols

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/simia-tech/go-pop3"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type POP3Test struct {
	input   string
	options TestOptions
}

//
// Run the test against the specified target.
//
func (s *POP3Test) RunTest(target string) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 110

	//
	// If the user specified a different port update to use it.
	//
	out := ParseArguments(s.input)
	if out["port"] != "" {
		port, err = strconv.Atoi(out["port"])
		if err != nil {
			return err
		}
	}

	//
	// Default to connecting to an IPv4-address
	//
	address := fmt.Sprintf("%s:%d", target, port)

	//
	// If we find a ":" we know it is an IPv6 address though
	//
	if strings.Contains(target, ":") {
		address = fmt.Sprintf("[%s]:%d", target, port)
	}

	//
	// Connect
	//
	c, err := pop3.Dial(address, pop3.UseTimeout(s.options.Timeout))
	if err != nil {
		return err
	}

	//
	// Did we get a username/password?  If so try to authenticate
	// with them
	//
	if (out["username"] != "") && (out["password"] != "") {
		err = c.Auth(out["username"], out["password"])
		if err != nil {
			return err
		}
	}

	//
	// Quit and return
	//
	c.Quit()
	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *POP3Test) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *POP3Test) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("pop3", func() ProtocolTest {
		return &POP3Test{}
	})
}
