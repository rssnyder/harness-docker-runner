package runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/harness/lite-engine/logstream"
	"github.com/sirupsen/logrus"
)

func getNudges() []logstream.Nudge {
	// <search-term> <resolution> <error-msg>
	return []logstream.Nudge{
		logstream.NewNudge("[Kk]illed", "Increase memory resources for the step", errors.New("out of memory")),
		logstream.NewNudge(".*git.* SSL certificate problem",
			"Set sslVerify to false in CI codebase properties", errors.New("SSL certificate error")),
		logstream.NewNudge("Cannot connect to the Docker daemon",
			"Setup dind if it's not running. If dind is running, privileged should be set to true",
			errors.New("could not connect to the docker daemon")),
	}
}

func getOutputVarCmd(outputVars []string, outputFile string) string {
	cmd := ""
	for _, o := range outputVars {
		cmd += fmt.Sprintf(";echo \"%s $%s\" >> %s", o, o, outputFile)
	}

	return cmd
}

// Fetches map of env variable and value from OutputFile.
// OutputFile stores all env variable and value
func fetchOutputVariables(outputFile string, out io.Writer) (map[string]string, error) {
	log := logrus.New()
	log.Out = out

	outputs := make(map[string]string)
	f, err := os.Open(outputFile)
	if err != nil {
		log.WithError(err).WithField("outputFile", outputFile).Errorln("failed to open output file")
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		sa := strings.Split(line, " ")
		if len(sa) < 2 { // nolint:gomnd
			log.WithField("variable", sa[0]).Warnln("output variable does not exist")
		} else {
			outputs[sa[0]] = line[len(sa[0])+1:]
		}
	}
	if err := s.Err(); err != nil {
		log.WithError(err).Errorln("failed to create scanner from output file")
		return nil, err
	}
	return outputs, nil
}
