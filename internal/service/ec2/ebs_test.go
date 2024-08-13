// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2EBSDefaultKMSKey_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			acctest.CtBasic: testAccEBSDefaultKMSKey_basic,
		},
		"DataSource": {
			acctest.CtBasic: testAccEBSDefaultKMSKeyDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
