// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataLake_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "S3_MANAGED_KEY"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "s3_bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_identitySerial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		"Basic":            testAccDataLake_Identity_basic,
		"ExistingResource": testAccDataLake_Identity_ExistingResource_basic,
		"RegionOverride":   testAccDataLake_Identity_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccDataLake_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.region", acctest.Region()),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("securitylake", "data-lake/default")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
			{
				ImportStateKind:         resource.ImportCommandWithID,
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},

			// TODO: Plannable import cannot be tested due to import diffs
		},
	})
}

func testAccDataLake_Identity_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_data_lake.test"

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region(), acctest.AlternateRegion())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_regionOverride(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "configuration.0.region", acctest.AlternateRegion()),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("securitylake", "data-lake/default")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},

			// TODO: Plannable import cannot be tested due to import diffs
		},
	})
}

func testAccDataLake_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecuritylake.ResourceDataLake, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataLake_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
			{
				Config: testAccDataLakeConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataLakeConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccDataLake_lifeCycle(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_lifeCycle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.days", "80"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.1.storage_class", "ONEZONE_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
			{
				Config: testAccDataLakeConfig_lifeCycleUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_metaStoreUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.test"

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
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_metaStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
			{
				Config: testAccDataLakeConfig_metaStoreUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager_updated", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_replication(t *testing.T) {
	ctx := acctest.Context(t)
	var datalake types.DataLakeResource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_securitylake_data_lake.region_2"

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t, acctest.Region(), acctest.AlternateRegion())
	})

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeConfig_replication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataLakeExists(ctx, t, resourceName, &datalake),
					resource.TestCheckResourceAttrPair(resourceName, "meta_store_manager_role_arn", "aws_iam_role.meta_store_manager", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.encryption_configuration.0.kms_key_id", "S3_MANAGED_KEY"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.days", "31"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.transition.0.storage_class", "STANDARD_IA"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.lifecycle_configuration.0.expiration.0.days", "300"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.replication_configuration.0.role_arn", "aws_iam_role.datalake_s3_replication", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.replication_configuration.0.regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.replication_configuration.0.regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"meta_store_manager_role_arn"},
			},
		},
	})
}

func testAccDataLake_Identity_ExistingResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_data_lake.test"

	t.Cleanup(func() {
		testAccDeleteGlueDatabases(ctx, t)
	})

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		CheckDestroy: testAccCheckDataLakeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccDataLakeConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccDataLakeConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccDataLakeConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrARN)),
				},
			},
		},
	})
}

func testAccCheckDataLakeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_data_lake" {
				continue
			}

			_, err := tfsecuritylake.FindDataLakeByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Data Lake %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataLakeExists(ctx context.Context, t *testing.T, n string, v *types.DataLakeResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindDataLakeByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccDataLakeConfigConfig_base_kmsKey = `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [{
    "Sid": "Enable IAM User Permissions",
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": "kms:*",
    "Resource": "*"
  }]
}
POLICY
}
`

func testAccDataLakeConfigConfig_base_regionOverride() string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  region = %[1]q

  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [{
    "Sid": "Enable IAM User Permissions",
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": "kms:*",
    "Resource": "*"
  }]
}
POLICY
}
`, acctest.AlternateRegion()))
}

const testAccDataLakeConfigConfig_base = `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "meta_store_manager" {
  name               = "AmazonSecurityLakeMetaStoreManagerV2"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowLambda",
    "Effect": "Allow",
    "Principal": {
      "Service": [
        "lambda.amazonaws.com"
      ]
    },
    "Action": "sts:AssumeRole"
  }]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "datalake" {
  role       = aws_iam_role.meta_store_manager.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonSecurityLakeMetastoreManager"
}

resource "aws_iam_role" "datalake_s3_replication" {
  name               = "AmazonSecurityLakeS3ReplicationRole"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "s3.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

# These are required in all configurations because the role stays registered with the Lake Formation Data Lake
# after the MetaStoreManager role is updated.
resource "aws_iam_role" "meta_store_manager_updated" {
  name               = "AmazonSecurityLakeMetaStoreManagerV1"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowLambda",
    "Effect": "Allow",
    "Principal": {
      "Service": [
        "lambda.amazonaws.com"
      ]
    },
    "Action": "sts:AssumeRole"
  }]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "datalake_updated" {
  role       = aws_iam_role.meta_store_manager_updated.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonSecurityLakeMetastoreManager"
}

resource "aws_iam_role_policy" "datalake_s3_replication" {
  name = "AmazonSecurityLakeS3ReplicationRolePolicy"
  role = aws_iam_role.datalake_s3_replication.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowReadS3ReplicationSetting",
      "Action": [
        "s3:ListBucket",
        "s3:GetReplicationConfiguration",
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersion",
        "s3:GetObjectVersionAcl",
        "s3:GetObjectVersionTagging",
        "s3:GetObjectRetention",
        "s3:GetObjectLegalHold"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*",
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:ResourceAccount": [
            "${data.aws_caller_identity.current.account_id}"
          ]
        }
      }
    },
    {
      "Sid": "AllowS3Replication",
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete",
        "s3:ReplicateTags"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:ResourceAccount": [
            "${data.aws_caller_identity.current.account_id}"
          ]
        }
      }
    }
  ]
}
POLICY
}
`

func testAccDataLakeConfig_basic() string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, acctest.Region()))
}

func testAccDataLakeConfig_regionOverride() string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base_regionOverride(),
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  region = %[1]q

  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, acctest.AlternateRegion()))
}

func testAccDataLakeConfig_tags1(tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[3]q
  }

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, tag1Key, tag1Value, acctest.Region()))
}

func testAccDataLakeConfig_tags2(tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfigConfig_base, fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[5]q
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, tag1Key, tag1Value, tag2Key, tag2Value, acctest.Region()))
}

func testAccDataLakeConfig_lifeCycle(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		testAccDataLakeConfigConfig_base_kmsKey,
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[2]q

    encryption_configuration {
      kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      transition {
        days          = 80
        storage_class = "ONEZONE_IA"
      }
      expiration {
        days = 300
      }
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_lifeCycleUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		testAccDataLakeConfigConfig_base_kmsKey,
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[2]q

    encryption_configuration {
      kms_key_id = aws_kms_key.test.id
    }

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_metaStore(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[2]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_metaStoreUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfigConfig_base,
		fmt.Sprintf(`
resource "aws_securitylake_data_lake" "test" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager_updated.arn

  configuration {
    region = %[2]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake_updated]
}
`, rName, acctest.Region()))
}

func testAccDataLakeConfig_replication(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_data_lake" "region_2" {
  meta_store_manager_role_arn = aws_iam_role.meta_store_manager.arn

  configuration {
    region = %[3]q

    lifecycle_configuration {
      transition {
        days          = 31
        storage_class = "STANDARD_IA"
      }
      expiration {
        days = 300
      }
    }

    replication_configuration {
      role_arn = aws_iam_role.datalake_s3_replication.arn
      regions  = [%[2]q]
    }
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.datalake, aws_iam_role_policy.datalake_s3_replication, aws_securitylake_data_lake.test]
}
`, rName, acctest.Region(), acctest.AlternateRegion()))
}
