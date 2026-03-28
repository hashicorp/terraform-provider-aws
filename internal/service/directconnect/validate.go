// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validConnectionBandWidth() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"1Gbps",
		"2Gbps",
		"5Gbps",
		"10Gbps",
		"25Gbps",
		"100Gbps",
		"400Gbps",
		"50Mbps",
		"100Mbps",
		"200Mbps",
		"300Mbps",
		"400Mbps",
		"500Mbps"}, false)
}

func validLongASN() schema.SchemaValidateFunc {
	return func(v any, k string) (ws []string, errors []error) {
		value, ok := v.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		asn, err := parseASN(value)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q (%q) must be a valid ASN in asplain or asdot format: %w", k, value, err))
			return
		}

		if asn < 1 || asn > 4294967294 {
			errors = append(errors, fmt.Errorf("%q (%q) must be between 1 and 4294967294", k, value))
		}
		return
	}
}

func parseASN(value string) (int64, error) {
	if strings.Contains(value, ".") {
		parts := strings.Split(value, ".")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid asdot format")
		}
		high, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		low, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, err
		}
		return (high << 16) + low, nil
	}
	return strconv.ParseInt(value, 10, 64)
}
