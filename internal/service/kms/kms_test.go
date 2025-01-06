// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMS_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"CustomKeyStore": {
			acctest.CtBasic:      testAccCustomKeyStore_basic,
			"update":             testAccCustomKeyStore_update,
			acctest.CtDisappears: testAccCustomKeyStore_disappears,
			"DataSource_basic":   testAccCustomKeyStoreDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
