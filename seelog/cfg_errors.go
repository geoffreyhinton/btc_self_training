package seelog

import (
	"errors"
)

var (
	nodeMustHaveChildrenError   = errors.New("Node must have children")
	nodeCannotHaveChildrenError = errors.New("Node cannot have children")
)

type unexpectedChildElementError struct {
	baseError
}

func newUnexpectedChildElementError(msg string) *unexpectedChildElementError {
	custmsg := "Unexpected child element: " + msg
	return &unexpectedChildElementError{baseError{message: custmsg}}
}

type missingArgumentError struct {
	baseError
}

func newMissingArgumentError(nodeName, attrName string) *missingArgumentError {
	custmsg := "Output '" + nodeName + "' has no '" + attrName + "' attribute"
	return &missingArgumentError{baseError{message: custmsg}}
}

type unexpectedAttributeError struct {
	baseError
}

func newUnexpectedAttributeError(nodeName, attr string) *unexpectedAttributeError {
	custmsg := nodeName + " has unexpected attribute: " + attr
	return &unexpectedAttributeError{baseError{message: custmsg}}
}
