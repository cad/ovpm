package errors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// CLIError error group indicates command related errors.
// Starting with error code 2xxx.
const CLIError = 2000

// ErrUnknownCLIError indicates an unknown cli error.
const ErrUnknownCLIError = 2001

// UnknownCLIError ...
func UnknownCLIError(e error) Error {
	err := Error{
		Message: fmt.Sprintf("unknown cli error: %v", e),
		Code:    ErrUnknownCLIError,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}
