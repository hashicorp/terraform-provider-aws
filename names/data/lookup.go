// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package data

import "fmt"

func LookupService(name string) (result ServiceRecord, err error) {
	serviceData, err := ReadAllServiceData()
	if err != nil {
		return result, fmt.Errorf("error reading service data: %s", err)
	}

	for _, s := range serviceData {
		if name == s.ProviderPackage() {
			return s, nil
		}
	}

	return result, fmt.Errorf("package not found: %s", name)
}
