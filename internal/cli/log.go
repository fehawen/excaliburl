package cli

import (
	"fmt"
	"os"
)

func Log(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[excaliburl] "+format+"\n", args...)
}

func Hint() {
	Log("try 'excaliburl -h' for more information")
}

func Die(code int, format string, args ...any) {
	Log(format, args...)

	if code == 2 {
		Hint()
	}

	os.Exit(code)
}
