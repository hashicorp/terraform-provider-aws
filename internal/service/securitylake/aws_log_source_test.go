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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.accounts.#"),
					acctest.CheckResourceAttrAccountID(resourceName, "source.0.accounts.0"),
					func(s *terraform.State) error {
						return resource.TestCheckTypeSetElemAttr(resourceName, "source.0.accounts.*", acctest.AccountID())(s)
					},
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "2.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLogSourceConfig_sourceVersion("2.0"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccAWSLogSource_sourceVersion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	var logSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_sourceVersion("1.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", acctest.Ct1),
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
			{
				Config: testAccAWSLogSourceConfig_sourceVersion("2.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "2.0"),
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
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.Region()),
					resource.TestCheckTypeSetElemAttr(resourceName, "source.0.regions.*", acctest.AlternateRegion()),
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
			testAccPreCheck(ctx, t)
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

func testAccAWSLogSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	resourceName2 := "aws_securitylake_aws_log_source.test2"
	var logSource, logSource2 types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, resourceName, &logSource),
					testAccCheckAWSLogSourceExists(ctx, resourceName2, &logSource2),

					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "2.0"),

					resource.TestCheckResourceAttr(resourceName2, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName2, "source.0.source_name", "S3_DATA"),
					resource.TestCheckResourceAttr(resourceName2, "source.0.source_version", "2.0"),
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
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`)
}

func testAccAWSLogSourceConfig_sourceVersion(version string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.name]
    source_name    = "ROUTE53"
    source_version = %[1]q
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, version))
}

func testAccAWSLogSourceConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccDataLakeConfig_replication(rName), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name, data.aws_region.alternate.name]
    source_name = "ROUTE53"
  }

  depends_on = [aws_securitylake_data_lake.test, aws_securitylake_data_lake.region_2]
}

data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}
`)
}

func testAccAWSLogSourceConfig_multiple() string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_securitylake_aws_log_source" "test2" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
    source_name = "S3_DATA"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`)
}
