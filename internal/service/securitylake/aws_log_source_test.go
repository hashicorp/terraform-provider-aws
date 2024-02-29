// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAWSLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	var logSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "1.0"),
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

func testAccAWSLogSource_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var logSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.Region()),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "1.0"),
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

func testAccAWSLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	var logSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceAWSLogSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLogSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_aws_log_source" {
				continue
			}

			_, err := tfsecuritylake.FindAWSLogSourceBySourceName(ctx, conn, types.AwsLogSourceName(rs.Primary.ID))

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake AWS Log Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSLogSourceExists(ctx context.Context, n string, v *types.AwsLogSourceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindAWSLogSourceBySourceName(ctx, conn, types.AwsLogSourceName(rs.Primary.ID))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSLogSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.test.account_id]
    regions        = [%[1]q]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}
`, acctest.Region()))
}

func testAccAWSLogSourceConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_replication(rName), fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.test.account_id]
    regions        = [%[1]q, %[2]q]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }

  depends_on = [aws_securitylake_data_lake.test, aws_securitylake_data_lake.region_2]
}
`, acctest.Region(), acctest.AlternateRegion()))
}
