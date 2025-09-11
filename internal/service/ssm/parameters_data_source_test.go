// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"
)

func TestAccSSMParametersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	allResourceName := "data.aws_ssm_parameters.test"
	filteredResourceName := "data.aws_ssm_parameters.filtered"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParametersDataSourceConfig_basic(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(allResourceName, "arns.#", "3"),
					resource.TestCheckResourceAttr(filteredResourceName, "arns.#", "2"),
				),
			},
		},
	})
}

func testAccParametersDataSourceConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test1" {
  name  = "/%[1]s/param-a"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "test2" {
  name  = "/%[1]s/param-b"
  type  = "String"
  value = "TestValueB"
}

resource "aws_ssm_parameter" "test3" {
  name  = "/%[2]s/param-c"
  type  = "String"
  value = "TestValueC"
}

data "aws_ssm_parameters" "test" {
  depends_on = [
    aws_ssm_parameter.test1,
    aws_ssm_parameter.test2,
    aws_ssm_parameter.test3,
  ]
}

data "aws_ssm_parameters" "filtered" {

  parameter_filter {
	key = "Name"
	option = "BeginsWith"
	values = ["/%[1]s/"]
  }

  depends_on = [
    aws_ssm_parameter.test1,
    aws_ssm_parameter.test2,
    aws_ssm_parameter.test3,
  ]
}
`, rName1, rName2)
}

func TestAccSSMParametersDataSource_ramShared(t *testing.T) {
	ctx := acctest.Context(t)

	sharedResourceName := "data.aws_ssm_parameters.test_shared"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccParametersDataSourceConfig_ramShared(rName1, rName2),
			},
			{
				Config: testAccParametersDataSourceConfig_ramShared(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(sharedResourceName, "arns.#", "2"),
				),
			},
		},
	})
}

func testAccParametersDataSourceConfig_ramShared(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_caller_identity.current.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ssm_parameter" "test1" {
  name  = "/%[1]s/param-a"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "test2" {
  provider = "awsalternate"

  name  = "/%[2]s/param-b"
  type  = "String"
  tier = "Advanced" #Required for RAM sharing
  value = "TestValueB"
}

resource "aws_ssm_parameter" "test3" {
  provider = "awsalternate"

  name  = "/%[2]s/param-c"
  type  = "String"
  tier = "Advanced" #Required for RAM sharing
  value = "TestValueC"
}

resource "aws_ram_resource_association" "test2" {
  provider = "awsalternate"

  resource_arn       = aws_ssm_parameter.test2.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_resource_association" "test3" {
  provider = "awsalternate"

  resource_arn       = aws_ssm_parameter.test3.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

data "aws_ssm_parameters" "test_shared" {
	shared = true
}
`, rName1, rName2))
}
