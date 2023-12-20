// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package data

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"io"
)

type ServiceRecord []string

func ReadAllServiceData() (results []ServiceRecord, err error) {
	reader := csv.NewReader(bytes.NewReader(namesData))
	// reader.ReuseRecord = true

	// Skip the header
	reader.Read()

	for {
		r, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil
		}
		results = append(results, ServiceRecord(r))
	}

	return
}

//go:embed names_data.csv
var namesData []byte
