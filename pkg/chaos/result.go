package chaos

type Result struct {
	Chaos *Chaos
	Err   error
}

// NewResult creates a new Result instance
func NewResult(chaos *Chaos, err error) *Result {
	return &Result{
		Chaos: chaos,
		Err:   err,
	}
}
