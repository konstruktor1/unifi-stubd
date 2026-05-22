// exitError carries CLI exit-code intent through ordinary Go errors.
// Validation, parse, and runtime failures can stay distinct without threading
// numeric exits through every call site.
package main

// exitError wraps a normal error with the process exit code the CLI should use.
type exitError struct {
	code int
	err  error
}

// Error returns the wrapped CLI error text.
func (e exitError) Error() string {
	return e.err.Error()
}

// Unwrap exposes the underlying error for callers using errors.Is/As.
func (e exitError) Unwrap() error {
	return e.err
}

// ExitCode returns the process status code associated with the error.
func (e exitError) ExitCode() int {
	return e.code
}

// withExitCode attaches CLI-oriented exit codes without changing the underlying
// error message.
func withExitCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return exitError{code: code, err: err}
}
