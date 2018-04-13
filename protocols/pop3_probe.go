package protocols

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
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
	// If the user specified a different port update it.
	//
	re := regexp.MustCompile("on\\s+port\\s+([0-9]+)")
	out := re.FindStringSubmatch(s.input)
	if len(out) == 2 {
		port, err = strconv.Atoi(out[1])
		if err != nil {
			return err
		}
	}

	//
	// Set an explicit timeout
	//
	d := net.Dialer{Timeout: s.options.Timeout}

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
	// Make the TCP connection.
	//
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return err
	}

	//
	// Read the banner.
	//
	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	conn.Close()

	if !strings.Contains(banner, "+OK") {
		return errors.New("Banner doesn't look like an rsync-banner")
	}

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