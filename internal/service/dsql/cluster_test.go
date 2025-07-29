// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdsql "github.com/hashicorp/terraform-provider-aws/internal/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// Because dsql is in preview, we need to skip the PreCheckPartitionHasService
			// acctest.PreCheckPartitionHasService(t, names.DSQLEndpointID)
			// PreCheck for the region configuration as long as DSQL is in preview
			acctest.PreCheckRegion(t, "us-east-1", "us-east-2")          //lintignore:AWSAT003
			acctest.PreCheckAlternateRegion(t, "us-east-2", "us-east-1") //lintignore:AWSAT003
			acctest.PreCheckThirdRegion(t, "us-west-2")                  //lintignore:AWSAT003
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// Because dsql is in preview, we need to skip the PreCheckPartitionHasService
			// acctest.PreCheckPartitionHasService(t, names.DSQLEndpointID)
			// PreCheck for the region configuration as long as DSQL is in preview
			acctest.PreCheckRegion(t, "us-east-1", "us-east-2")          //lintignore:AWSAT003
			acctest.PreCheckAlternateRegion(t, "us-east-2", "us-east-1") //lintignore:AWSAT003
			acctest.PreCheckThirdRegion(t, "us-west-2")                  //lintignore:AWSAT003
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdsql.ResourceCluster, resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// Because dsql is in preview, we need to skip the PreCheckPartitionHasService
			// acctest.PreCheckPartitionHasService(t, names.DSQLEndpointID)
			// PreCheck for the region configuration as long as DSQL is in preview
			acctest.PreCheckRegion(t, "us-east-1", "us-east-2")          //lintignore:AWSAT003
			acctest.PreCheckAlternateRegion(t, "us-east-2", "us-east-1") //lintignore:AWSAT003
			acctest.PreCheckThirdRegion(t, "us-west-2")                  //lintignore:AWSAT003
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
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
				Config: testAccClusterConfig_basic(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
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

func TestAccDSQLCluster_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// Because dsql is in preview, we need to skip the PreCheckPartitionHasService
			// acctest.PreCheckPartitionHasService(t, names.DSQLEndpointID)
			// PreCheck for the region configuration as long as DSQL is in preview
			acctest.PreCheckRegion(t, "us-east-1", "us-east-2")          //lintignore:AWSAT003
			acctest.PreCheckAlternateRegion(t, "us-east-2", "us-east-1") //lintignore:AWSAT003
			acctest.PreCheckThirdRegion(t, "us-west-2")                  //lintignore:AWSAT003
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_cmk(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
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
					testAccCheckClusterExists(ctx, resourceName, &cluster),
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

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dsql_cluster" {
				continue
			}

			_, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

			if tfresource.NotFound(err) {
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

func testAccCheckClusterExists(ctx context.Context, n string, v *dsql.GetClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

		output, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

	input := dsql.ListClustersInput{}
	_, err := conn.ListClusters(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_basic(deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = %[1]t
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
