package encoder

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fehawen/excaliburl/internal/result"
)

type JSONL struct {
	w io.Writer
}

func New(w io.Writer) *JSONL {
	return &JSONL{
		w: w,
	}
}

func (j *JSONL) Write(r result.Result) error {
	out, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(j.w, string(out))
	return err
}
