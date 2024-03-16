// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package namevaluesfiltersv2

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"github.com/aws/aws-sdk-go-v2/aws"
	imagebuildertypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
)

// []*SERVICE.Filter handling

// ImagebuilderFilters returns imagebuilder service filters.
func (filters NameValuesFilters) ImagebuilderFilters() []imagebuildertypes.Filter {
	m := filters.Map()

	if len(m) == 0 {
		return nil
	}

	result := make([]imagebuildertypes.Filter, 0, len(m))

	for k, v := range m {
		filter := imagebuildertypes.Filter{
			Name:   aws.String(k),
			Values: v,
		}

		result = append(result, filter)
	}

	return result
}
