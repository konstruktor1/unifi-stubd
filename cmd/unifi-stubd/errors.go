// exitError carries CLI exit-code intent through ordinary Go errors.
// Validation, parse, and runtime failures can stay distinct without threading
// numeric exits through every call site.
package main

type exitError struct {
	code int
	err  error
}

func (e exitError) Error() string {
	return e.err.Error()
}

func (e exitError) Unwrap() error {
	return e.err
}

func (e exitError) ExitCode() int {
	return e.code
}

func withExitCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return exitError{code: code, err: err}
}
