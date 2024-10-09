// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package v1

import (
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
)

type NameValuesFilters struct {
	namevaluesfilters.NameValuesFilters
}

func New(i interface{}) NameValuesFilters {
	return NameValuesFilters{NameValuesFilters: namevaluesfilters.New(i)}
}
