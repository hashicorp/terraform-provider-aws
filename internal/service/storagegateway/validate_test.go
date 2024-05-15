// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
)

func TestValid4ByteASN(t *testing.T) {
	t.Parallel()

	validAsns := []string{
		acctest.Ct0,
		acctest.Ct1,
		"65534",
		"65535",
		"4294967294",
		"4294967295",
	}
	for _, v := range validAsns {
		_, errors := tfstoragegateway.Valid4ByteASN(v, "bgp_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"-1",
		"ABCDEFG",
		"",
		"4294967296",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := tfstoragegateway.Valid4ByteASN(v, "bgp_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}
