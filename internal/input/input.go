package input

import "os"

type Input struct {
	Path  string
	Stdin bool
}

func StdinAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) == 0
}
