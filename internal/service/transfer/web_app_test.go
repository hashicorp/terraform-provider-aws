// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferWebApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_endpoint"), knownvalue.StringRegexp(regexache.MustCompile(`^https:\/\/.*.aws$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("transfer", regexache.MustCompile(`webapp/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("web_app_endpoint_policy"), tfknownvalue.StringExact(awstypes.WebAppEndpointPolicyStandard)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("web_app_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"provisioned": knownvalue.Int64Exact(1),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
		},
	})
}

func TestAccTransferWebApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tftransfer.ResourceWebApp, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTransferWebApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccWebAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccTransferWebApp_webAppUnits(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_webAppUnits(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("web_app_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"provisioned": knownvalue.Int64Exact(2),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppConfig_webAppUnits(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("web_app_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"provisioned": knownvalue.Int64Exact(4),
						}),
					})),
				},
			},
		},
	})
}

func TestAccTransferWebApp_accessEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_accessEndPoint(rName, "https://example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_endpoint"), knownvalue.StringExact("https://example.com")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppConfig_accessEndPoint(rName, "https://example.net"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("access_endpoint"), knownvalue.StringExact("https://example.net")),
				},
			},
		},
	})
}

func TestAccTransferWebApp_VPC(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedWebApp
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_VPC(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName,
						tfjsonpath.New("endpoint_details").AtSliceIndex(0).AtMapKey("vpc").AtSliceIndex(0).AtMapKey(names.AttrVPCID),
						"aws_vpc.test",
						tfjsonpath.New(names.AttrID),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_app_id",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "web_app_id"),
			},
			{
				Config: testAccWebAppConfig_VPC(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						resourceName,
						tfjsonpath.New("endpoint_details").AtSliceIndex(0).AtMapKey("vpc").AtSliceIndex(0).AtMapKey(names.AttrVPCID),
						"aws_vpc.test",
						tfjsonpath.New(names.AttrID),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccCheckWebAppDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_web_app" {
				continue
			}

			_, err := tftransfer.FindWebAppByID(ctx, conn, rs.Primary.Attributes["web_app_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Web App %s still exists", rs.Primary.Attributes["web_app_id"])
		}

		return nil
	}
}

func testAccCheckWebAppExists(ctx context.Context, t *testing.T, n string, v *awstypes.DescribedWebApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		resp, err := tftransfer.FindWebAppByID(ctx, conn, rs.Primary.Attributes["web_app_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccWebAppConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "assume_role_transfer" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
      "sts:SetContext"
    ]
    principals {
      type        = "Service"
      identifiers = ["transfer.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_transfer.json
}

data "aws_iam_policy_document" "web_app_identity_bearer" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetDataAccess",
      "s3:ListCallerAccessGrants",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:access-grants/*"
    ]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ListAccessGrantsInstances"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
}

resource "aws_iam_role_policy" "web_app_identity_bearer" {
  policy = data.aws_iam_policy_document.web_app_identity_bearer.json
  role   = aws_iam_role.test.name
}
`, rName)
}

func testAccWebAppConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_base(rName), `
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
}
`)
}

func testAccWebAppConfig_webAppUnits(rName string, webAppUnitsProvisioned int) string {
	return acctest.ConfigCompose(testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  web_app_units {
    provisioned = %[2]d
  }
}
`, rName, webAppUnitsProvisioned))
}

func testAccWebAppConfig_accessEndPoint(rName, accessEndPoint string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  access_endpoint = %[2]q
}
`, rName, accessEndPoint))
}

func testAccWebAppConfig_VPC(rName string, subnetIndex int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		testAccWebAppConfig_base(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  endpoint_details {
    vpc {
      vpc_id             = aws_vpc.test.id
      subnet_ids         = [aws_subnet.test[%[2]d].id]
      security_group_ids = [aws_security_group.test.id]
    }
  }
}
`, rName, subnetIndex))
}

func testAccWebAppConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccWebAppConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}
