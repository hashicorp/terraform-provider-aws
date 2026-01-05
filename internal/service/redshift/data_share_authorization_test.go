// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftDataShareAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_data_share_authorization.test"
	callerIdentityDataSourceName := "data.aws_caller_identity.current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataShareAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataShareAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataShareAuthorizationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "consumer_identifier", callerIdentityDataSourceName, names.AttrAccountID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "data_share_arn", "redshift", regexache.MustCompile(`datashare:+.`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "producer_arn", "redshift-serverless", regexache.MustCompile(`namespace/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftDataShareAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_data_share_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataShareAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataShareAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataShareAuthorizationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceDataShareAuthorization, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataShareAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_data_share_authorization" {
				continue
			}

			_, err := tfredshift.FindDataShareAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_share_arn"], rs.Primary.Attributes["consumer_identifier"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Data Share Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataShareAuthorizationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		_, err := tfredshift.FindDataShareAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_share_arn"], rs.Primary.Attributes["consumer_identifier"])

		return err
	}
}

func testAccDataShareAuthorizationConfigBase(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  db_name        = "test"
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftdata_statement" "test_create" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = "CREATE DATASHARE tfacctest;"
}
`, rName),
		// Split this resource into a string literal so the terraform `format` function
		// interpolates properly
		`
resource "aws_redshiftdata_statement" "test_grant_usage" {
  depends_on     = [aws_redshiftdata_statement.test_create]
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = format("GRANT USAGE ON DATASHARE tfacctest TO ACCOUNT '%s';", data.aws_caller_identity.current.account_id)
}
`)
}

func testAccDataShareAuthorizationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataShareAuthorizationConfigBase(rName),
		`
resource "aws_redshift_data_share_authorization" "test" {
  depends_on = [aws_redshiftdata_statement.test_grant_usage]

  # Data share ARN is not returned from the GRANT USAGE statement, so must be
  # composed manually.
  # Ref: https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonredshift.html#amazonredshift-resources-for-iam-policies
  data_share_arn = format("arn:%s:redshift:%s:%s:datashare:%s/%s",
    data.aws_partition.current.id,
    data.aws_region.current.region,
    data.aws_caller_identity.current.account_id,
    aws_redshiftserverless_namespace.test.namespace_id,
    "tfacctest",
  )

  consumer_identifier = data.aws_caller_identity.current.account_id
}
`)
}
