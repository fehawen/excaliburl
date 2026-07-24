package input

import (
	"io/fs"
	"path/filepath"
)

func WalkDir(root string) ([]Input, error) {
	var inputs []Input

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		inputs = append(inputs, Input{
			Path: abs,
		})

		return nil
	})

	return inputs, err
}
