package custom_errors

import "fmt"

type ArgError struct {
	Arg     string
	Message string
}

type QuorumWriteError struct {
	Message string
	W       int
	N       int
}
type QuorumReadError struct {
	Message string
	R       int
	N       int
}

func (e *QuorumWriteError) Error() string {
	return fmt.Sprintf("%s w %v n %v", e.Message, e.W, e.N)
}

func (e *QuorumReadError) Error() string {
	return fmt.Sprintf("%s r %v n %v", e.Message, e.R, e.N)
}
func (e *ArgError) Error() string {
	return fmt.Sprintf("%s - %s", e.Arg, e.Message)
}
