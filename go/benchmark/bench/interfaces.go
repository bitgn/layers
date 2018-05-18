package bench

// Launcher - a common interface for running benchmarks
type Launcher interface {
	Describe() *Description
	Exec(r uint64) error
}

type Description struct {
	Code  string
	Name  string
	Setup string
}
