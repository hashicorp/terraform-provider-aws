package common

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadAllCSVData(filename string) ([][]string, error) {
	f, err := os.Open(filename)

	if err != nil {
		return nil, fmt.Errorf("opening file (%s): %w", filename, err)
	}

	defer f.Close()

	return csv.NewReader(f).ReadAll()
}
