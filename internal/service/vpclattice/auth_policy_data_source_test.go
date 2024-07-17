// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeAuthPolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_auth_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthPolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrPolicy, regexache.MustCompile(`"Action":"*"`)),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_identifier", "aws_vpclattice_service.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccAuthPolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_vpclattice_auth_policy" "test" {
  resource_identifier = aws_vpclattice_auth_policy.test.resource_identifier
}

resource "aws_vpclattice_service" "test" {
  name               = %[1]q
  auth_type          = "AWS_IAM"
  custom_domain_name = "example.com"
}

resource "aws_vpclattice_auth_policy" "test" {
  resource_identifier = aws_vpclattice_service.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "*"
      Effect    = "Allow"
      Principal = "*"
      Resource  = "*"
      Condition = {
        StringNotEqualsIgnoreCase = {
          "aws:PrincipalType" = "anonymous"
        }
      }
    }]
  })
}
`, rName)
}
