package input

import "io"

func isText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	sample := data

	if len(sample) > 8192 {
		sample = sample[:8192]
	}

	var control int

	for _, b := range sample {
		if b == 0 {
			return false
		}

		if b < 32 &&
			b != '\n' &&
			b != '\r' &&
			b != '\t' {
			control++
		}
	}

	ratio := float64(control) / float64(len(sample))

	return ratio < 0.10
}

func ReadText(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if !isText(data) {
		return nil, nil
	}

	return data, nil
}
