// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validConnectionBandWidth() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"1Gbps",
		"2Gbps",
		"5Gbps",
		"10Gbps",
		"100Gbps",
		"50Mbps",
		"100Mbps",
		"200Mbps",
		"300Mbps",
		"400Mbps",
		"500Mbps"}, false)
}
