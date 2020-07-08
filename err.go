package protoerror

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type protoValidationError interface {
	Field() string
	Reason() string
	Cause() error
	Key() bool
	Error() string
}

type ValidationError struct {
	Field  string
	Reason string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Reason)
}

func (ve ValidationError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, ve.Error())
}

func FormatValidationError(err error) error {
	ve, ok := err.(protoValidationError)
	if !ok {
		return err
	}

	if cause := ve.Cause(); cause != nil {
		if causeRaw, ok := cause.(protoValidationError); ok {
			cause := FormatValidationError(causeRaw)
			if subve, ok := cause.(*ValidationError); ok {
				return &ValidationError{
					Field:  ve.Field() + "." + subve.Field,
					Reason: subve.Reason,
				}
			}
			return cause
		}
	}

	return &ValidationError{
		Field:  ve.Field(),
		Reason: ve.Reason(),
	}
}
