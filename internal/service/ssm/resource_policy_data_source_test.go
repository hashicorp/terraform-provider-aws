// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMResourcePolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssm_resource_policy.test"
	resourceName := "aws_ssm_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy_hash", resourceName, "policy_hash"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccSSMResourcePolicyDataSource_byPolicyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssm_resource_policy.test"
	resourceName := "aws_ssm_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyDataSourceConfig_byPolicyID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy_id", resourceName, "policy_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccResourcePolicyDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourcePolicyConfig_basic(rName), `
data "aws_ssm_resource_policy" "test" {
  resource_arn = aws_ssm_resource_policy.test.resource_arn
}
`)
}

func testAccResourcePolicyDataSourceConfig_byPolicyID(rName string) string {
	return acctest.ConfigCompose(testAccResourcePolicyConfig_basic(rName), `
data "aws_ssm_resource_policy" "test" {
  resource_arn = aws_ssm_resource_policy.test.resource_arn
  policy_id    = aws_ssm_resource_policy.test.policy_id
}
`)
}
