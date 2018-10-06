package errors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// SystemError error group indicates system related errors.
// Starting with error code 1xxx.
const SystemError = 1000

// ErrUnknownSysError indicates an unknown system error.
const ErrUnknownSysError = 1001

// UnknownSysError ...
func UnknownSysError(e error) Error {
	err := Error{
		Message: fmt.Sprintf("unknown system error: %v", e),
		Code:    ErrUnknownSysError,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrUnknownGRPCError indicates an unknown gRPC error.
const ErrUnknownGRPCError = 1002

// UnknownGRPCError ...
func UnknownGRPCError(e error) Error {
	err := Error{
		Message: fmt.Sprintf("grpc error: %v", e),
		Code:    ErrUnknownGRPCError,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrUnknownFileIOError indicates an unknown File IO error.
const ErrUnknownFileIOError = 1003

// UnknownFileIOError ...
func UnknownFileIOError(e error) Error {
	err := Error{
		Message: fmt.Sprintf("file io error: %v", e),
		Code:    ErrUnknownFileIOError,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}
