package domainerror

type ValidationError struct {
	err error
}

func NewErrValidation(err error) *ValidationError {
	return &ValidationError{
		err: err,
	}
}

func (e *ValidationError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}
