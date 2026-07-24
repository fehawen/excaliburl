package processor

import "github.com/fehawen/excaliburl/internal/result"

type Emitter func(result.Result)

type Processor interface {
	Process(data []byte, file string, emit Emitter)
}
