// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"
)

func TestValidConnectionBandWidth(t *testing.T) {
	t.Parallel()

	validBandwidths := []string{
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
		"500Mbps",
	}
	for _, v := range validBandwidths {
		_, errors := validConnectionBandWidth()(v, "bandwidth")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid bandwidth: %q", v, errors)
		}
	}

	invalidBandwidths := []string{
		"1Tbps",
		"10GBpS",
		"42Mbps",
		"0",
		"???",
		"a lot",
	}
	for _, v := range invalidBandwidths {
		_, errors := validConnectionBandWidth()(v, "bandwidth")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid bandwidth", v)
		}
	}
}
