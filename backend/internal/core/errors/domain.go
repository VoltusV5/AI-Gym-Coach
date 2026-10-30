package core_errors

type DomainError struct {
	Code    string
	Message string
	Status  int
	Cause   error
}

func (e *DomainError) Error() string { return e.Code }
func (e *DomainError) Unwrap() error { return e.Cause }
