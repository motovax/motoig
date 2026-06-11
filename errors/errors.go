// Package errors defines typed errors for motoig.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for errors.Is checks.
var (
	ErrAuthentication = errors.New("motoig: authentication failed")
	ErrSessionExpired = errors.New("motoig: session expired")
	ErrLoginRequired  = errors.New("motoig: login required")
	ErrInstagramAPI   = errors.New("motoig: instagram api error")
	ErrNetwork        = errors.New("motoig: network error")
	ErrParsing        = errors.New("motoig: parsing error")
	ErrValidation     = errors.New("motoig: validation error")
	ErrThrottled      = errors.New("motoig: rate limited")
	ErrChallenge      = errors.New("motoig: challenge required")
	ErrTwoFactor      = errors.New("motoig: two-factor required")
	ErrBadPassword    = errors.New("motoig: bad password")
	ErrPrivateAccount = errors.New("motoig: private account")
	ErrMediaNotFound  = errors.New("motoig: media not found")
	ErrUserNotFound   = errors.New("motoig: user not found")
	ErrStoryNotFound  = errors.New("motoig: story not found")
	ErrDirectNotFound = errors.New("motoig: direct thread not found")
	ErrCommentsDisabled = errors.New("motoig: comments disabled")
	ErrProxyBlocked   = errors.New("motoig: proxy address is blocked")
	ErrSentryBlock    = errors.New("motoig: sentry block")
	ErrFeedback       = errors.New("motoig: feedback required")
)

// Error is the base error type with optional code and details.
type Error struct {
	Op      string
	Message string
	Code    int
	Err     error
	Details map[string]any
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	if e.Code != 0 {
		return fmt.Sprintf("%s: %s (code %d)", e.Op, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func New(op, message string) *Error {
	return &Error{Op: op, Message: message}
}

func Wrap(op, message string, err error) *Error {
	return &Error{Op: op, Message: message, Err: err}
}

func WithCode(op, message string, code int) *Error {
	return &Error{Op: op, Message: message, Code: code}
}
