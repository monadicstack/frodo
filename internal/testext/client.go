package testext

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/monadicstack/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

// ExternalClientSuite is a test suite that validates the behavior of service clients generated for languages other
// than Go. It relies on you having a "runner" executable in the client's target language that runs one of the test
// cases we want to test and parses stdout of that program to determine whether the test should pass/fail.
//
// The suite contains the logic to fire up a local instance of the gateway for the client to hit on the desired port
// as well as the ability to shut it down after the test. You can then analyze each line of stdout to determine if
// each interaction behaved as expected and write your Go assertions based on that. There will be more detail in the
// Frodo architecture markdown docs as to how this all works.
type ExternalClientSuite struct {
	suite.Suite
	Server *http.Server
}

// StartService runs an instance of the gateway on the given address/port (e.g. ":9100"). Once this function returns
// you can fire off your client runner to hit this service.
func (suite *ExternalClientSuite) StartService(addr string, handler http.Handler) {
	suite.Server = &http.Server{Addr: addr, Handler: handler}
	go func() {
		err := suite.Server.ListenAndServe()
		if err != http.ErrServerClosed {
			suite.Fail("Could not start service gateway", err)
		}
	}()
}

// StopService shuts down the running service gateway for the current test. It's safe so that if the server never
// started or shut down unexpectedly during the test, this will silently fail since the work is still done.
func (suite *ExternalClientSuite) StopService() {
	if suite.Server != nil {
		_ = suite.Server.Shutdown(context.Background())
	}
}

// ExpectPass analyzes line N (zero-based) of the client runner's output and asserts that it was a successful
// interaction (e.g. output was "OK { ... some json ... }"). It will decode the right-hand-side JSON onto your
// 'out' parameter so that you can also run custom checks to fire after the decoding is complete to ensure that
// the actual output value is what you expect.
//
//     response := calc.AddResponse{}
//     suite.ExpectPass(output, 0, &response, func() {
//         suite.Equal(12, response.Result)
//     })
func (suite *ExternalClientSuite) ExpectPass(results ClientTestResults, lineIndex int, out interface{}, additionalChecks ...func()) {
	result := results[lineIndex]
	suite.Require().Equal(true, result.Pass, "Client Line %d: Call failed when it should have passed", lineIndex)
	suite.Require().NoError(result.Decode(out), "Client Line %d: Unable to unmarshal JSON", lineIndex)
	for _, additionalCheck := range additionalChecks {
		additionalCheck()
	}
}

// ExpectFail analyzes line N (zero-based) of the client runner's output and asserts that it was a failed
// interaction (e.g. output was "FAIL { ... some json ... }"). It will decode the right-hand-side JSON onto your
// 'out' parameter so that you can also run custom checks to fire after the decoding is complete to ensure that
// the actual output error is what you expect.
//
//     err := errors.RPCError{}
//     suite.ExpectFail(output, 0, &response, func() {
//         suite.Contains(err.Message, "divide by zero")
//     })
func (suite *ExternalClientSuite) ExpectFail(results ClientTestResults, lineIndex int, out interface{}, additionChecks ...func()) {
	result := results[lineIndex]
	suite.Require().Equal(false, result.Pass, "Client Line %d: Call passed when it should have failed", lineIndex)
	suite.Require().NoError(result.Decode(out), "Client Line %d: Unable to unmarshal JSON", lineIndex)
	for _, additionCheck := range additionChecks {
		additionCheck()
	}
}

// ExpectFailStatus analyzes line N (zero-based) of the client runner's output and asserts that it was a failed
// interaction (e.g. output was "FAIL { ... some json ... }"). It assumes that the right-hand JSON conforms to Frodo's
// RPCError such that there's a 'status' and 'message' field and that the status is equivalent to the one you supply.
// This is a convenience on top of ExpectFail to reduce verbosity.
//
//     // Assumes the first case failed w/ a 403 status, the second w/ a 502, and the last with a 500.
//     suite.ExpectFailStatus(output, 0, 403)
//     suite.ExpectFailStatus(output, 1, 502)
//     suite.ExpectFailStatus(output, 2, 500)
func (suite *ExternalClientSuite) ExpectFailStatus(results ClientTestResults, lineIndex int, status int) {
	err := errors.RPCError{}
	suite.ExpectFail(results, lineIndex, &err, func() {
		suite.Require().Equal(status, err.Status())
	})
}

// RunClientTest executes the language-specific runner to execute a single test case in that language. The result
// of the runner's execution are written to stdout w/ each interaction on a separate line. This will delegate to
// ParseClientTestOutput() to turn it into an easily workable value that you can hit with your Go assertions.
func RunClientTest(executable string, args ...string) (ClientTestResults, error) {
	stdout, err := exec.Command(executable, args...).Output()
	if err != nil {
		return nil, err
	}
	return ParseClientTestOutput(stdout), nil
}

// ParseClientTestOutput accepts the entire stdout of RunClientTest and parses each line to determine how each
// interaction in the test case behaved. Here is a sample output of a runner that performed 5 client calls to
// the backend service; 3 that passed and 2 that failed.
//
//     OK {"result": 5}
//     FAIL {"message": "divide by zero", "status": 400}
//     OK {"result": 3.14}
//     OK {"result": 10, "remainder": 2}
//     FAIL {"message": "overflow", "status": 400}
//
// All language runners should output in this format for this to work. It's a convention that allows us to build
// assertions in Go regardless of how the target language does its work.
func ParseClientTestOutput(stdout []byte) ClientTestResults {
	var results ClientTestResults

	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		pass := strings.HasPrefix(line, "OK")
		output := strings.TrimPrefix(strings.TrimPrefix(line, "OK "), "FAIL")
		results = append(results, ClientTestResult{Pass: pass, Output: []byte(output)})
	}
	return results
}

// ClientTestResults encapsulates all of the output lines from a client test runner.
type ClientTestResults []ClientTestResult

// ClientTestResult decodes a single output line of stdout from a client test runner. It parses "OK {...}"
// or "FAIL {...}" and makes it easier for your test assertion code to work with.
type ClientTestResult struct {
	// Pass is true when the line started with "OK", false otherwise.
	Pass bool
	// Output is the raw characters output to stdout for this interaction.
	Output []byte
}

// Decode overlays this output line's JSON on top of the 'out' parameter (i.e. the stuff after OK/FAIL).
func (res ClientTestResult) Decode(out interface{}) error {
	return json.Unmarshal(res.Output, out)
}

// String just regurgitates the original output line.
func (res ClientTestResult) String() string {
	return string(res.Output)
}
