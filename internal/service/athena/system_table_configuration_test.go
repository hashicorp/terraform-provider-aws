// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	awstypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaSystemTableConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		acctest.CtBasic:      testAccSystemTableConfiguration_basic,
		acctest.CtDisappears: testAccSystemTableConfiguration_disappears,
		"alreadyExists":      testAccSystemTableConfiguration_alreadyExists,
		"kmsKey":             testAccSystemTableConfiguration_kmsKey,
		"kmsKeyChange":       testAccSystemTableConfiguration_kmsKeyChange,
		"tags":               testAccAthenaSystemTableConfiguration_tagsSerial,
		"Identity":           testAccAthenaSystemTableConfiguration_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSystemTableConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_system_table_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccSystemTableConfigurationPreCheck(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckSystemTableConfigurationDestroy(ctx, t),
			testAccCheckSystemTablesRetained(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemTableConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
					testAccCheckSystemTablesEnabled(ctx, t),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "athena", "system-table/{table_name}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTableName), tfknownvalue.StringExact(awstypes.SystemTableNameAll)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        string(awstypes.SystemTableNameAll),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
			},
		},
	})
}

func testAccSystemTableConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_system_table_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccSystemTableConfigurationPreCheck(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemTableConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemTableConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfathena.ResourceSystemTableConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccSystemTableConfiguration_alreadyExists(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_system_table_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccSystemTableConfigurationPreCheck(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemTableConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				PreConfig:   testAccCreateSystemTableConfiguration(ctx, t),
				Config:      testAccSystemTableConfigurationConfig_basic(),
				ExpectError: regexache.MustCompile(`import\s+it\s+with\s+the\s+ID\s+"ALL"`),
			},
			{
				Config:             testAccSystemTableConfigurationConfig_basic(),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      string(awstypes.SystemTableNameAll),
				ImportStatePersist: true,
			},
			{
				Config: testAccSystemTableConfigurationConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccSystemTableConfiguration_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_system_table_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccSystemTableConfigurationPreCheck(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemTableConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccSystemTableConfigurationConfig_kmsKeyARN(fmt.Sprintf("arn:%s:kms:%s:123456789012:alias/example", acctest.Partition(), acctest.Region())),
				PlanOnly:    true,
				ExpectError: regexache.MustCompile(`not\s+an\s+alias\s+ARN`),
			},
			{
				Config: testAccSystemTableConfigurationConfig_kmsKey(rName, "aws_kms_key.test[0].arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), "aws_kms_key.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        string(awstypes.SystemTableNameAll),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrTableName,
			},
			{
				Config: testAccSystemTableConfigurationConfig_kmsKey(rName, "aws_kms_key.test[1].arn"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ExpectError: regexache.MustCompile(`different\s+encryption`),
			},
			{
				Config: testAccSystemTableConfigurationConfig_kmsKey(rName, "aws_kms_key.test[0].arn"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccSystemTableConfiguration_kmsKeyChange(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_system_table_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccSystemTableConfigurationPreCheck(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemTableConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemTableConfigurationConfig_kmsKey(rName, "aws_kms_key.test[0].arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), "aws_kms_key.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				PreConfig: testAccDeleteSystemTables(ctx, t),
				Config:    testAccSystemTableConfigurationConfig_kmsKey(rName, "aws_kms_key.test[1].arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), "aws_kms_key.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccSystemTableConfigurationConfig_kmsKey(rName, "null"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ExpectError: regexache.MustCompile(`different\s+encryption`),
			},
			{
				PreConfig: testAccDeleteSystemTables(ctx, t),
				Config:    testAccSystemTableConfigurationConfig_kmsKey(rName, "null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSystemTableConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrKMSKeyARN), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccSystemTableConfigurationPreCheck(ctx context.Context, t *testing.T, region string) {
	ctx = acctest.NewTestResourceContext(ctx, "aws_athena_system_table_configuration", region)
	client := acctest.ProviderMeta(ctx, t)

	_, err := tfathena.FindSystemTableConfigurationByName(ctx, client.AthenaClient(ctx), awstypes.SystemTableNameAll)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err == nil {
		t.Skipf("skipping acceptance testing: an Athena system table configuration exists in %s", region)
	}
	if !retry.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	tables, err := tfathena.FindSystemTables(ctx, client)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
	if len(tables) > 0 {
		t.Skipf("skipping acceptance testing: %d Athena system tables exist in namespace metadata of table bucket aws-athena in %s", len(tables), region)
	}

	// The test context is canceled before cleanups run.
	ctx = context.WithoutCancel(ctx)
	t.Cleanup(func() {
		if err := tfathena.ResetSystemTables(ctx, client); err != nil {
			t.Errorf("resetting Athena system tables in %s: %s", region, err)
		}
	})
}

func testAccCheckSystemTableConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		_, err := tfathena.FindSystemTableConfigurationByName(ctx, conn, awstypes.SystemTableName(rs.Primary.Attributes[names.AttrTableName]))

		return err
	}
}

func testAccCheckSystemTableConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_system_table_configuration" {
				continue
			}

			name := rs.Primary.Attributes[names.AttrTableName]
			_, err := tfathena.FindSystemTableConfigurationByName(ctx, conn, awstypes.SystemTableName(name))
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Athena System Table Configuration %s still exists", name)
		}

		return nil
	}
}

func testAccCheckSystemTablesEnabled(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		var input athena.ListSystemTableStatusesInput
		var n int
		pages := athena.NewListSystemTableStatusesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				return err
			}

			for _, v := range page.Items {
				if !aws.ToBool(v.Enabled) {
					return fmt.Errorf("Athena system table %s is not enabled", v.TableName)
				}
				n++
			}
		}

		if n == 0 {
			return fmt.Errorf("no Athena system table statuses")
		}

		return nil
	}
}

func testAccCheckSystemTablesRetained(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tables, err := tfathena.FindSystemTables(ctx, acctest.ProviderMeta(ctx, t))
		if err != nil {
			return err
		}

		if len(tables) == 0 {
			return fmt.Errorf("Athena system tables were deleted with the configuration")
		}

		return nil
	}
}

func testAccCreateSystemTableConfiguration(ctx context.Context, t *testing.T) func() {
	return func() {
		conn := acctest.ProviderMeta(ctx, t).AthenaClient(ctx)

		input := athena.CreateSystemTableConfigurationInput{
			TableName: awstypes.SystemTableNameAll,
		}
		if _, err := conn.CreateSystemTableConfiguration(ctx, &input); err != nil {
			t.Fatalf("creating Athena System Table Configuration: %s", err)
		}
	}
}

func testAccDeleteSystemTables(ctx context.Context, t *testing.T) func() {
	return func() {
		if err := tfathena.DeleteSystemTables(ctx, acctest.ProviderMeta(ctx, t)); err != nil {
			t.Fatalf("deleting Athena system tables: %s", err)
		}
	}
}

func testAccSystemTableConfigurationConfig_basic() string {
	return `
resource "aws_athena_system_table_configuration" "test" {}
`
}

func testAccSystemTableConfigurationConfig_kmsKeyARN(kmsKeyARN string) string {
	return fmt.Sprintf(`
resource "aws_athena_system_table_configuration" "test" {
  kms_key_arn = %[1]q
}
`, kmsKeyARN)
}

func testAccSystemTableConfigurationConfig_kmsKey(rName, kmsKeyARNExpression string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = "Root"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

  statement {
    sid       = "AthenaSystemTables"
    actions   = ["kms:GenerateDataKey", "kms:Decrypt", "kms:DescribeKey"]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["systemtables.athena.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    sid       = "S3TablesMaintenance"
    actions   = ["kms:GenerateDataKey", "kms:Decrypt", "kms:DescribeKey"]
    resources = ["*"]

    principals {
      type        = "Service"
      identifiers = ["maintenance.s3tables.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values   = ["arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:bucket/aws-athena/*"]
    }
  }
}

resource "aws_kms_key" "test" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.test.json
}

resource "aws_athena_system_table_configuration" "test" {
  kms_key_arn = %[2]s
}
`, rName, kmsKeyARNExpression)
}
