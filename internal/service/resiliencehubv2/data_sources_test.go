// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2PolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_policy.test"
	dataSourceName := "data.aws_resiliencehubv2_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "availability_slo.0.target", "99.9"),
				),
			},
		},
	})
}

func TestAccResilienceHubV2SystemDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_system.test"
	dataSourceName := "data.aws_resiliencehubv2_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccResilienceHubV2ServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_service.test"
	dataSourceName := "data.aws_resiliencehubv2_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "permission_model.0.invoker_role_name", "AWSResilienceHubAssessmentRole"),
				),
			},
		},
	})
}

func testAccPolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_policy" "test" {
  name = %[1]q

  availability_slo {
    target = 99.9
  }
}

data "aws_resiliencehubv2_policy" "test" {
  arn = aws_resiliencehubv2_policy.test.arn
}
`, rName)
}

func testAccSystemDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_system" "test" {
  name = %[1]q
}

data "aws_resiliencehubv2_system" "test" {
  arn = aws_resiliencehubv2_system.test.arn
}
`, rName)
}

func testAccServiceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_policy" "test" {
  name = "%[1]s-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  name    = %[1]q
  regions = ["us-west-2"]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

data "aws_resiliencehubv2_service" "test" {
  arn = aws_resiliencehubv2_service.test.arn
}
`, rName)
}
