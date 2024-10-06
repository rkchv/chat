package err

import (
	"errors"
)

type commonError struct {
	msg  string
	code Code
}

func New(msg string, code Code) *commonError {
	return &commonError{msg, code}
}

func NewFromError(err error, code Code) *commonError {
	return &commonError{err.Error(), code}
}

func (r *commonError) Error() string {
	return r.msg
}

func (r *commonError) Code() Code {
	return r.code
}

func IsCommonError(err error) bool {
	var ce *commonError
	return errors.As(err, &ce)
}

func GetCommonError(err error) *commonError {
	var ce *commonError
	if !errors.As(err, &ce) {
		return nil
	}

	return ce
}
