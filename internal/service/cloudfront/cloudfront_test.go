// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testAccRegionProviderConfig is the Terraform provider configuration for CloudFront region testing
//
// Testing CloudFront assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRegionProviderConfig() string {
	switch acctest.Partition() {
	case names.StandardPartitionID:
		return acctest.ConfigRegionalProvider(names.USEast1RegionID)
	case names.ChinaPartitionID:
		return acctest.ConfigRegionalProvider(names.CNNorthwest1RegionID)
	default:
		return acctest.ConfigRegionalProvider(acctest.Region())
	}
}
