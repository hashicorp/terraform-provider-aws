// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdsql "github.com/hashicorp/terraform-provider-aws/internal/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("dsql", regexache.MustCompile(`cluster/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("encryption_details"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"encryption_status": tfknownvalue.StringExact(awstypes.EncryptionStatusEnabled),
							"encryption_type":   tfknownvalue.StringExact(awstypes.EncryptionTypeAwsOwnedKmsKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrForceDestroy), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kms_encryption_key"), knownvalue.StringExact("AWS_OWNED_KMS_KEY")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_service_name"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrIdentifier),
				ImportStateVerifyIdentifierAttribute: names.AttrIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccDSQLCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdsql.ResourceCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSQLCluster_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deletionProtection(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrIdentifier),
				ImportStateVerifyIdentifierAttribute: names.AttrIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccClusterConfig_deletionProtection(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccDSQLCluster_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_forceDestroy(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrForceDestroy), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccDSQLCluster_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_cmk(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("encryption_details"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"encryption_status": tfknownvalue.StringExact(awstypes.EncryptionStatusEnabled),
							"encryption_type":   tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerManagedKmsKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kms_encryption_key"), tfknownvalue.RegionalARNRegexp("kms", regexache.MustCompile(`key/.+$`))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrIdentifier),
				ImportStateVerifyIdentifierAttribute: names.AttrIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccClusterConfig_awsOwnedKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cluster),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("encryption_details"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"encryption_status": tfknownvalue.StringExact(awstypes.EncryptionStatusEnabled),
							"encryption_type":   tfknownvalue.StringExact(awstypes.EncryptionTypeAwsOwnedKmsKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kms_encryption_key"), knownvalue.StringExact("AWS_OWNED_KMS_KEY")),
				},
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dsql_cluster" {
				continue
			}

			_, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Aurora DSQL Cluster %s still exists", rs.Primary.Attributes[names.AttrIdentifier])
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, t *testing.T, n string, v *dsql.GetClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

		output, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

	input := dsql.ListClustersInput{}
	_, err := conn.ListClusters(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_basic() string {
	return `
resource "aws_dsql_cluster" "test" {
}
`
}

func testAccClusterConfig_deletionProtection(deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = %[1]t
}
`, deletionProtection)
}

func testAccClusterConfig_forceDestroy(deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = %[1]t
  force_destroy               = true
}
`, deletionProtection)
}

func testAccClusterConfig_baseEncryptionDetails(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Sid    = "Enable dsql IAM User Permissions"
        Effect = "Allow"
        Principal = {
          Service = "dsql.amazonaws.com"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          "AWS" : "*"
        }
        Action   = "kms:*"
        Resource = "*"
      }
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccClusterConfig_cmk(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_baseEncryptionDetails(rName), `
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  kms_encryption_key          = aws_kms_key.test.arn

  depends_on = [
    aws_kms_key_policy.test
  ]
}
`)
}

func testAccClusterConfig_awsOwnedKey(rName string) string { // nosemgrep:ci.aws-in-func-name
	return acctest.ConfigCompose(testAccClusterConfig_baseEncryptionDetails(rName), `
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  kms_encryption_key          = "AWS_OWNED_KMS_KEY"
}
`)
}
