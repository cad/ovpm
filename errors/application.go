package errors

import (
	"fmt"
	"net/url"

	"github.com/Sirupsen/logrus"
)

// ApplicationError error group indicates application related errors.
// Starting with error code 3xxx.
const ApplicationError = 3000

// ErrUnknownApplicationError indicates an unknown application error.
const ErrUnknownApplicationError = 3001

// UnknownApplicationError ...
func UnknownApplicationError(e error) Error {
	err := Error{
		Message: fmt.Sprintf("unknown application error: %v", e),
		Code:    ErrUnknownApplicationError,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrMustBeLoopbackURL indicates that given url does not resolv to a known looback ip addr.
const ErrMustBeLoopbackURL = 3002

// MustBeLoopbackURL ...
func MustBeLoopbackURL(url *url.URL) Error {
	err := Error{
		Message: "url must resolve to a known looback ip addr",
		Code:    ErrMustBeLoopbackURL,
		Args: map[string]interface{}{
			"url": url.String(),
		},
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrBadURL indicates that given url string can not be parsed.
const ErrBadURL = 3003

// BadURL ...
func BadURL(urlStr string, e error) Error {
	err := Error{
		Message: fmt.Sprintf("url string can not be parsed: %v", e),
		Code:    ErrBadURL,
		Args: map[string]interface{}{
			"url": urlStr,
		},
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrEmptyValue indicates that given value is empty.
const ErrEmptyValue = 3004

// EmptyValue ...
func EmptyValue(key string, value interface{}) Error {
	err := Error{
		Message: fmt.Sprintf("value is empty: %v", value),
		Code:    ErrEmptyValue,
		Args: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrNotIPv4 indicates that given value is not an IPv4.
const ErrNotIPv4 = 3005

// NotIPv4 ...
func NotIPv4(str string) Error {
	err := Error{
		Message: fmt.Sprintf("'%s' is not an IPv4 address", str),
		Code:    ErrNotIPv4,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrConflictingDemands indicates that users demands are conflicting with each other.
const ErrConflictingDemands = 3006

// ConflictingDemands ...
func ConflictingDemands(msg string) Error {
	err := Error{
		Message: msg,
		Code:    ErrConflictingDemands,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrNotHostname indicates that given value is not an hostname.
const ErrNotHostname = 3007

// NotHostname ...
func NotHostname(str string) Error {
	err := Error{
		Message: fmt.Sprintf("'%s' is not a valid host name", str),
		Code:    ErrNotHostname,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrNotCIDR indicates that given value is not a CIDR.
const ErrNotCIDR = 3008

// NotCIDR ...
func NotCIDR(str string) Error {
	err := Error{
		Message: fmt.Sprintf("'%s' is not a valid CIDR", str),
		Code:    ErrNotCIDR,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrInvalidPort indicates that given value is not a valid port number.
const ErrInvalidPort = 3009

// InvalidPort ...
func InvalidPort(str string) Error {
	err := Error{
		Message: fmt.Sprintf("'%s' is not a valid port number", str),
		Code:    ErrInvalidPort,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrUnconfirmed indicates a UI confirmation dialog is cancelled by the user.
const ErrUnconfirmed = 3010

// Unconfirmed ...
func Unconfirmed(str string) Error {
	err := Error{
		Message: fmt.Sprintf("confirmation failed: '%s'", str),
		Code:    ErrUnconfirmed,
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}

// ErrNotValidNetworkType indicates that supplied network type is invalid.
const ErrNotValidNetworkType = 3011

// NotValidNetworkType ...
func NotValidNetworkType(key string, value interface{}) Error {
	err := Error{
		Message: fmt.Sprintf("invalid network type: %v", value),
		Code:    ErrNotValidNetworkType,
		Args: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}
	logrus.WithFields(logrus.Fields(err.Args)).Error(err)
	return err
}
