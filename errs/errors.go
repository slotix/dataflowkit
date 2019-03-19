// Dataflow kit - Errs
//
// Copyright © 2017-2018 Slotix s.r.o. <dm@slotix.sk>
//
//
// All rights reserved. Use of this source code is governed
// by the BSD 3-Clause License license.

package errs

import "fmt"

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

const (
	ErrNoParts                  = "no parts found"
	ErrNoSelectors              = "no selectors found"
	ErrEmptyResults             = "empty results"
	ErrNoCommonAncestor         = "no common ancestor for selectors found"
	ErrNoPartOrSelectorProvided = "no selector/name provided for %s"
)

//BadPayload error is returned if Payload is invalid 400
type BadPayload struct {
	ErrText string
}

func (e BadPayload) Error() string {
	return e.ErrText
}

func (e BadPayload) Status() int {
	return 400
}

// ErrStorageResult represent storage results reader errors
type ErrStorageResult struct {
	Err string
}

// Exported Storage Result errors
const (
	EOF      = "End of payload results"
	NextPage = "Next page results"
	NoKey    = "Key %s not found"
)

func (e *ErrStorageResult) Error() string {
	return e.Err
}

// Cancel error inform about operation canceled by user
type Cancel struct {
}

func (c Cancel) Error() string {
	return "Operation canceled."
}

type ParseError struct {
	URL string
	Err error
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s : %s", e.URL, e.Err.Error())
}
