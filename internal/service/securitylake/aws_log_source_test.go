// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAWSLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_aws_log_source.test"
	var logSource types.AwsLogSourceConfiguration

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.accounts.#"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "source.0.accounts.0"),
					func(s *terraform.State) error {
						return resource.TestCheckTypeSetElemAttr(resourceName, "source.0.accounts.*", acctest.AccountID(ctx))(s)
					},
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", "1"),
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

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_sourceVersion("1.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
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
			{
				Config: testAccAWSLogSourceConfig_sourceVersion("2.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var logSource types.AwsLogSourceConfiguration

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region(), acctest.AlternateRegion())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.regions.#", "2"),
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

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecuritylake.ResourceAWSLogSource, resourceName),
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

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAWSLogSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLogSourceConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLogSourceExists(ctx, t, resourceName, &logSource),
					testAccCheckAWSLogSourceExists(ctx, t, resourceName2, &logSource2),

					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.source_version", "2.0"),

					resource.TestCheckResourceAttr(resourceName2, "source.#", "1"),
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

func testAccCheckAWSLogSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_aws_log_source" {
				continue
			}

			_, err := tfsecuritylake.FindAWSLogSourceBySourceName(ctx, conn, types.AwsLogSourceName(rs.Primary.ID))

			if retry.NotFound(err) {
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

func testAccCheckAWSLogSourceExists(ctx context.Context, t *testing.T, n string, v *types.AwsLogSourceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindAWSLogSourceBySourceName(ctx, conn, types.AwsLogSourceName(rs.Primary.ID))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAWSLogSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`)
}

func testAccAWSLogSourceConfig_sourceVersion(version string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.current.account_id]
    regions        = [data.aws_region.current.region]
    source_name    = "ROUTE53"
    source_version = %[1]q
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`, version))
}

func testAccAWSLogSourceConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), testAccDataLakeConfig_replication(rName), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region, data.aws_region.alternate.region]
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
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), `
resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
    source_name = "ROUTE53"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_securitylake_aws_log_source" "test2" {
  source {
    accounts    = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
    source_name = "S3_DATA"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

data "aws_region" "current" {}
`)
}
