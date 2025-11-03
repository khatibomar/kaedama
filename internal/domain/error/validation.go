package domainError

type ErrValidation struct {
	err error
}

func NewErrValidation(err error) *ErrValidation {
	return &ErrValidation{
		err: err,
	}
}

func (e *ErrValidation) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}
