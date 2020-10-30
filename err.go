package protoerror

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
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

type validator interface {
	Validate() error
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				return nil, FormatValidationError(err)
			}
		}
		res, err := handler(ctx, req)
		if err != nil {
			if _, ok := status.FromError(err); !ok {
				return nil, status.Error(codes.Internal, "Internal Error")
			}
			return nil, err
		}
		return res, nil
	}
}
